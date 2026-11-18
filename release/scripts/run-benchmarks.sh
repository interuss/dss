#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=release/scripts/common.sh
source "$SCRIPT_DIR/common.sh"

BENCHMARKER_DIR="$RELEASE_DIR/tests/benchmarker"

MONITORING_IMAGE="${MONITORING_IMAGE:-interuss/monitoring:latest}"
DUMMY_OAUTH_IMAGE="${DUMMY_OAUTH_IMAGE:-interuss/dummy-oauth}"
TEST_NETWORK="dss-release-benchmarker-net"
OAUTH_CONTAINER="dss-release-benchmarker-oauth"
OAUTH_PORT=8085
OAUTH_URL_IN_NET="http://$OAUTH_CONTAINER:$OAUTH_PORT/token"

cleanup() {
    local running
    running=$(docker ps --format '{{.Names}}' 2>/dev/null | grep -E '^dss-release-bench-' || true)
    # shellcheck disable=SC2015,SC2086
    [[ -n "$running" ]] && docker stop $running >/dev/null 2>&1 || true
    docker stop "$OAUTH_CONTAINER" >/dev/null 2>&1 || true
    docker rm   "$OAUTH_CONTAINER" >/dev/null 2>&1 || true
    docker network rm "$TEST_NETWORK" >/dev/null 2>&1 || true
}

# === 1. Install Docker ===
section "Install Docker (if missing)"
ensure_docker

trap cleanup EXIT INT TERM

# === 2. Start test network + dummy-oauth ===
section "Start test network + dummy-oauth"
docker rm -f "$OAUTH_CONTAINER" >/dev/null 2>&1 || true
docker network create "$TEST_NETWORK" >/dev/null 2>&1 || true
ok "network $TEST_NETWORK"

docker run -d --name "$OAUTH_CONTAINER" \
    --network "$TEST_NETWORK" \
    -p "$OAUTH_PORT:$OAUTH_PORT" \
    -v "$REPO_ROOT/build/test-certs:/var/test-certs:ro" \
    "$DUMMY_OAUTH_IMAGE" \
    -private_key_file /var/test-certs/auth2.key >/dev/null

for i in $(seq 1 30); do
    if curl -fs "http://localhost:$OAUTH_PORT/token?intended_audience=-&scope=-" >/dev/null 2>&1; then
        break
    fi
    sleep 1
    (( i == 30 )) && die "dummy-oauth not reachable on :$OAUTH_PORT after 30s"
done
ok "dummy-oauth listening on :$OAUTH_PORT"

# === 3. Benchmarks: pools in parallel, configs of a pool sequentially ===
mapfile -t pools < <(find "$BENCHMARKER_DIR" -maxdepth 1 -type f -name '*_*.jsonnet' -printf '%f\n' \
                     | sed -E 's/^.*_([^_]+)\.jsonnet$/\1/' | sort -u)
(( ${#pools[@]} > 0 )) || die "no *_<pool>.jsonnet configs in $BENCHMARKER_DIR"

section "Benchmarks - ${#pools[@]} pool(s) in parallel: ${pools[*]}"
START=$(date +%s)
for pool in "${pools[@]}"; do
    mapfile -t configs < <(find "$BENCHMARKER_DIR" -maxdepth 1 -type f -name "*_$pool.jsonnet" -printf '%f\n' \
                           | sed -E 's/\.jsonnet$//' | sort)
    printf '  %s▶%s  %-8s %s\n' "$BLUE" "$RESET" "$pool" "${configs[*]}"
    START_OF[$pool]=$(date +%s)
    (
        for cfg in "${configs[@]}"; do
            echo "===== $cfg ($(date -u +%FT%TZ))"
            docker run --rm \
                --name "dss-release-bench-$cfg" \
                --network "$TEST_NETWORK" \
                -u "$(id -u):$(id -g)" \
                -e PYTHONUNBUFFERED=1 \
                -e AUTH_SPEC="DummyOAuth($OAUTH_URL_IN_NET,benchmarker)" \
                -v "$BENCHMARKER_DIR:/benchmarker" \
                -w /app/monitoring/benchmarker \
                "$MONITORING_IMAGE" \
                uv run benchmark.py \
                    --config "file:///benchmarker/$cfg.jsonnet" \
                    --output-dir "/benchmarker/output/$cfg" \
                || exit 1
        done
    ) >"$LOG_DIR/$pool.benchmarker.log" 2>&1 &
    PID_OF[$pool]=$!
done
echo

failed=0
wait_jobs benchmarker "${pools[@]}" || failed=1
printf '  %stotal: %s%s\n' "$DIM" "$(fmt_dur "$(( $(date +%s) - START ))")" "$RESET"

# === Summary ===
section "Summary"
printf '  outputs : %s/output/<config>/\n' "$BENCHMARKER_DIR"
printf '  logs    : %s/<pool>.benchmarker.log\n\n' "$LOG_DIR"
if (( failed != 0 )); then
    printf '  %sFAILED%s\n' "$BOLD$RED" "$RESET"
    exit 1
fi
printf '  %sALL DONE%s\n' "$BOLD$GREEN" "$RESET"
