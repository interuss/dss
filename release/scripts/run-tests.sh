#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=release/scripts/common.sh
source "$SCRIPT_DIR/common.sh"

TESTS_DIR="$RELEASE_DIR/tests"
OUTPUT_DIR="$RELEASE_DIR/uss_qualifier_output"
CACHE_DIR="$RELEASE_DIR/.templates_cache"

MONITORING_IMAGE="${MONITORING_IMAGE:-interuss/monitoring:latest}"
DUMMY_OAUTH_IMAGE="${DUMMY_OAUTH_IMAGE:-release-dummy-oauth}"
TEST_NETWORK="dss-release-test-net"
OAUTH_CONTAINER="dss-release-dummy-oauth"
OAUTH_PORT=8085
OAUTH_URL_IN_NET="http://$OAUTH_CONTAINER:$OAUTH_PORT/token"

mkdir -p "$OUTPUT_DIR" "$CACHE_DIR"
rm -f "$LOG_DIR"/*.prober.log "$LOG_DIR"/*.prober.junit.xml "$LOG_DIR"/*.uss_qualifier.log 2>/dev/null || true

cleanup() {
    local running
    running=$(docker ps --format '{{.Names}}' 2>/dev/null | grep -E '^dss-release-(prober|qual)-' || true)
    # shellcheck disable=SC2015,SC2086
    [[ -n "$running" ]] && docker stop $running >/dev/null 2>&1 || true
    docker stop "$OAUTH_CONTAINER" >/dev/null 2>&1 || true
    docker rm   "$OAUTH_CONTAINER" >/dev/null 2>&1 || true
    docker network rm "$TEST_NETWORK" >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM

# === 1. Resolve app_hostnames ===
section "Resolve app_hostnames"
declare -A HOST=()
for name in "${CLUSTERS[@]}"; do
    host=$(get_app_hostname "$RELEASE_INFRA_DIR/$name.tfvars")
    [[ -n "$host" ]] || die "app_hostname missing in $name.tfvars"
    HOST[$name]="$host"
    printf '  %s•%s  %-32s %s\n' "$DIM" "$RESET" "$name" "$host"
done

# === 2. Ensure dummy-oauth image exists ===
section "Build dummy-oauth image (if missing)"
if ! docker image inspect "$DUMMY_OAUTH_IMAGE" >/dev/null 2>&1; then
    info "building from $REPO_ROOT/cmds/dummy-oauth/Dockerfile ..."
    ( cd "$REPO_ROOT" && docker build -f cmds/dummy-oauth/Dockerfile -t "$DUMMY_OAUTH_IMAGE" . ) \
        >"$LOG_DIR/dummy-oauth-build.log" 2>&1
    ok "built $DUMMY_OAUTH_IMAGE"
else
    ok "$DUMMY_OAUTH_IMAGE already present"
fi

# === 3. Start test network + dummy-oauth ===
section "Start test network + dummy-oauth"
docker rm -f "$OAUTH_CONTAINER" >/dev/null 2>&1 || true
docker network create "$TEST_NETWORK" >/dev/null 2>&1 || true
ok "network $TEST_NETWORK"

docker run -d --name "$OAUTH_CONTAINER" \
    --network "$TEST_NETWORK" \
    -p "$OAUTH_PORT:$OAUTH_PORT" \
    "$DUMMY_OAUTH_IMAGE" \
    -private_key_file /var/test-certs/auth2.key >/dev/null

# Wait until the token endpoint answers.
for i in $(seq 1 30); do
    if curl -fs "http://localhost:$OAUTH_PORT/token?intended_audience=-&scope=-" >/dev/null 2>&1; then
        break
    fi
    sleep 1
    (( i == 30 )) && die "dummy-oauth not reachable on :$OAUTH_PORT after 30s"
done
ok "dummy-oauth listening on :$OAUTH_PORT"

# === Phase 1: prober sequentially against each cluster ===
section "Phase 1 - prober against ${#CLUSTERS[@]} endpoints (sequential)"
phase1_failed=0
P1_START=$(date +%s)
for name in "${CLUSTERS[@]}"; do
    endpoint="https://${HOST[$name]}"
    log="$LOG_DIR/$name.prober.log"
    junit="$LOG_DIR/$name.prober.junit.xml"
    : > "$junit"   # docker -v <file>:<file> requires the host file to exist
    printf '  %s→%s  %-32s %s\n' "$BLUE" "$RESET" "$name" "$endpoint"
    start_ts=$(date +%s)
    if docker run --rm \
            --name "dss-release-prober-$name" \
            --network "$TEST_NETWORK" \
            -v "$junit:/app/test_result" \
            -w /app/monitoring/prober \
            "$MONITORING_IMAGE" \
            pytest . -rsx \
                --junitxml=/app/test_result \
                --dss-endpoint "$endpoint" \
                --rid-auth    "DummyOAuth($OAUTH_URL_IN_NET,sub=fake_uss)" \
                --rid-v2-auth "DummyOAuth($OAUTH_URL_IN_NET,sub=fake_uss)" \
                --scd-auth1   "DummyOAuth($OAUTH_URL_IN_NET,sub=fake_uss)" \
                --scd-auth2   "DummyOAuth($OAUTH_URL_IN_NET,sub=fake_uss2)" \
                --scd-api-version 1.0.0 \
            >"$log" 2>&1; then
        printf '    %s✓%s  prober ok in %s\n' "$GREEN" "$RESET" "$(fmt_dur "$(( $(date +%s) - start_ts ))")"
    else
        printf '    %s✗%s  prober FAILED in %s - log: %s\n' "$RED" "$RESET" "$(fmt_dur "$(( $(date +%s) - start_ts ))")" "$log"
        phase1_failed=1
    fi
done
printf '  %sphase 1 total: %s%s\n' "$DIM" "$(fmt_dur "$(( $(date +%s) - P1_START ))")" "$RESET"

# === Phase 2: uss_qualifier configs in parallel ===
section "Phase 2 - uss_qualifier configs (parallel)"
mapfile -t configs < <(find "$TESTS_DIR" -maxdepth 1 -type f \( -name '*.yaml' -o -name '*.yml' \) -printf '%f\n' \
                       | sed -E 's/\.(yaml|yml)$//' | sort)
if (( ${#configs[@]} == 0 )); then
    warn "no *.yaml configs in $TESTS_DIR"
    exit "$phase1_failed"
fi
printf '  found %d config(s): %s\n\n' "${#configs[@]}" "${configs[*]}"

P2_START=$(date +%s)
for cfg in "${configs[@]}"; do
    out="$OUTPUT_DIR/$cfg"
    mkdir -p "$out"
    START_OF[$cfg]=$(date +%s)
    (
        docker run --rm \
            --name "dss-release-qual-$cfg" \
            --network "$TEST_NETWORK" \
            -u "$(id -u):$(id -g)" \
            -e PYTHONBUFFERED=1 \
            -e AUTH_SPEC="DummyOAuth($OAUTH_URL_IN_NET,uss_qualifier)" \
            -e AUTH_SPEC_2="DummyOAuth($OAUTH_URL_IN_NET,uss_qualifier_2)" \
            -v "$TESTS_DIR:/app/monitoring/uss_qualifier/configurations/personal" \
            -v "$out:/app/monitoring/uss_qualifier/output" \
            -v "$CACHE_DIR:/app/monitoring/uss_qualifier/.templates_cache" \
            -w /app/monitoring/uss_qualifier \
            "$MONITORING_IMAGE" \
            uv run main.py --config "configurations.personal.$cfg"
    ) >"$LOG_DIR/$cfg.uss_qualifier.log" 2>&1 &
    PID_OF[$cfg]=$!
    printf '  %s▶%s  %-48s started (pid %d)\n' "$BLUE" "$RESET" "configurations.personal.$cfg" "${PID_OF[$cfg]}"
done
echo

phase2_failed=0
wait_jobs uss_qualifier "${configs[@]}" || phase2_failed=1
printf '  %sphase 2 total: %s%s\n' "$DIM" "$(fmt_dur "$(( $(date +%s) - P2_START ))")" "$RESET"

# === Summary ===
section "Summary"
printf '  prober junits  : %s/*.prober.junit.xml\n' "$LOG_DIR"
printf '  uss_qualifier  : %s/<config>/\n' "$OUTPUT_DIR"
printf '  logs           : %s\n\n' "$LOG_DIR"
if (( phase1_failed != 0 || phase2_failed != 0 )); then
    printf '  %sFAILED%s - phase1=%d phase2=%d\n' "$BOLD$RED" "$RESET" "$phase1_failed" "$phase2_failed"
    exit 1
fi
printf '  %sALL PASSED%s\n' "$BOLD$GREEN" "$RESET"
