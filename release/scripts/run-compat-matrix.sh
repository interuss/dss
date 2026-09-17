#!/usr/bin/env bash

set -euo pipefail

VERSIONS=(v0.20.2 v0.21.1 v0.22.0 v0.23.0)

MONITORING_IMAGE="interuss/monitoring:v0.34.0"

# This should be removed when we stop testing v0.20.2.
# Improvements landed in v0.21.x are exercised by monitoring images newer than
# v0.24.0, so an older image is needed to check compatibility with v0.20.2.
V0_20_2_MONITORING_IMAGE="interuss/monitoring:v0.24.0"

###############################################################################

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=release/scripts/common.sh
source "$SCRIPT_DIR/common.sh"


declare -A VIDX
for i in "${!VERSIONS[@]}"; do VIDX["${VERSIONS[$i]}"]="$i"; done

DSS_REGISTRY="docker.io/interuss/dss"
OAUTH_IMAGE="interuss-local/dummy-oauth"
CRDB_IMAGE="cockroachdb/cockroach:v24.1.3"

NET="dss-compat"
COMPOSE=(docker compose -f "$SCRIPT_DIR/../compat/docker-compose.yaml")

QUAL_OUT="$RELEASE_DIR/uss_qualifier_output/compat"
mkdir -p "$LOG_DIR" "$QUAL_OUT"

declare -A VERDICT DETAIL

datastore_flag() {
    case "$1" in
        v0.20.2) echo "--cockroach_host" ;;
        *)       echo "--datastore_host" ;;
    esac
}

public_endpoint_flag() {
    case "$1" in
        v0.20.2) echo "" ;;
        *)       echo "-public_endpoint http://$2:8082" ;;
    esac
}

has_aux() {
    [[ "$1" != "v0.20.2" ]]
}

wait_healthy() {
    local svc="$1" cid state
    cid=$("${COMPOSE[@]}" ps -q "$svc" 2>/dev/null) || return 1
    [[ -n "$cid" ]] || return 1
    for _ in $(seq 1 90); do
        state=$(docker inspect -f '{{.State.Status}}:{{if .State.Health}}{{.State.Health.Status}}{{end}}' "$cid" 2>/dev/null || echo gone:)
        case "$state" in
            running:healthy)   return 0 ;;
            running:unhealthy) return 1 ;;
            exited:*|dead:*|gone:*) return 1 ;;
        esac
        sleep 2
    done
    return 1
}

teardown() {
    "${COMPOSE[@]}" down -v --remove-orphans >/dev/null 2>&1 || true
}
on_interrupt() {
    trap - EXIT INT TERM
    docker rm -f dss-compat-prober dss-compat-qualifier >/dev/null 2>&1 || true
    teardown
    exit 130
}
trap teardown EXIT
trap on_interrupt INT TERM

run_prober() {
    local endpoint="$1" log="$2" image="$3"
    docker run --rm \
        --name dss-compat-prober \
        --network "$NET" \
        -w /app/monitoring/prober \
        "$image" \
        pytest . -rsx \
            --dss-endpoint "http://$endpoint:8082" \
            --rid-auth "DummyOAuth(http://oauth:8085/token,sub=fake_uss)" \
            --rid-v2-auth "DummyOAuth(http://oauth:8085/token,sub=fake_uss)" \
            --scd-auth1 "DummyOAuth(http://oauth:8085/token,sub=fake_uss)" \
            --scd-auth2 "DummyOAuth(http://oauth:8085/token,sub=fake_uss2)" \
            --scd-api-version 1.0.0 \
        >"$log" 2>&1
}

run_qualifier() {
    local out="$1" log="$2" image="$3"
    rm -rf "$out"
    mkdir -p "$out"
    docker run --rm \
        --name dss-compat-qualifier \
        --network "$NET" \
        -u "$(id -u):$(id -g)" \
        -e AUTH_SPEC='DummyOAuth(http://oauth:8085/token,uss_qualifier)' \
        -e AUTH_SPEC_2='DummyOAuth(http://oauth:8085/token,uss_qualifier_2)' \
        -v "$SCRIPT_DIR/../compat/qualifier_config.yaml:/app/monitoring/uss_qualifier/compat_config.yaml:ro" \
        -v "$out:/app/monitoring/uss_qualifier/output" \
        -w /app/monitoring/uss_qualifier \
        "$image" \
        python main.py --config compat_config \
        >"$log" 2>&1
}


run_combo() {
    local a="$1" b="$2" key="$1|$2"
    local slug="${a}_${b}"

    export DSS_IMAGE_A="$DSS_REGISTRY:$a"
    export DSS_IMAGE_B="$DSS_REGISTRY:$b"
    export DATASTORE_FLAG_A DATASTORE_FLAG_B PUBLIC_ENDPOINT_A PUBLIC_ENDPOINT_B COMPOSE_PROFILES
    DATASTORE_FLAG_A="$(datastore_flag "$a")"
    DATASTORE_FLAG_B="$(datastore_flag "$b")"
    PUBLIC_ENDPOINT_A="$(public_endpoint_flag "$a" core-service-a)"
    PUBLIC_ENDPOINT_B="$(public_endpoint_flag "$b" core-service-b)"
    COMPOSE_PROFILES=""
    if has_aux "$b"; then
        COMPOSE_PROFILES="with-aux"
    fi

    "${COMPOSE[@]}" down -v --remove-orphans >/dev/null 2>&1 || true
    if ! "${COMPOSE[@]}" up -d >"$LOG_DIR/$slug.compose.log" 2>&1 \
        || ! wait_healthy oauth || ! wait_healthy core-service-a || ! wait_healthy core-service-b; then
        "${COMPOSE[@]}" logs >>"$LOG_DIR/$slug.compose.log" 2>&1 || true
        VERDICT[$key]="INFRA"
        DETAIL[$key]="stack did not come up"
        return
    fi

    local pa=0 pb=0 q=0 image="$MONITORING_IMAGE"
    if [[ "$a" == "v0.20.2" || "$b" == "v0.20.2" ]]; then
        image="$V0_20_2_MONITORING_IMAGE"
    fi
    run_prober core-service-a "$LOG_DIR/$slug.prober-a.log" "$image" || pa=1
    run_prober core-service-b "$LOG_DIR/$slug.prober-b.log" "$image" || pb=1

    run_qualifier "$QUAL_OUT/$slug" "$LOG_DIR/$slug.qualifier.log" "$image" || q=1

    DETAIL[$key]="prober-a=$( ((pa)) && echo fail || echo ok ) prober-b=$( ((pb)) && echo fail || echo ok ) qualifier=$( ((q)) && echo fail || echo ok )"

    if (( pa )) || (( pb )) || (( q )); then
        VERDICT[$key]="FAIL"
    else
        VERDICT[$key]="PASS"
    fi
}

symbol() {
    case "$1" in
        PASS)  printf '%s✓%s' "$GREEN" "$RESET" ;;
        FAIL)  printf '%s✗%s' "$RED" "$RESET" ;;
        INFRA) printf '%s⨯%s' "$RED" "$RESET" ;;
        *)     printf ' ' ;;
    esac
}

emoji() {
    case "$1" in
        PASS)  echo '✅' ;;
        FAIL)  echo '❌' ;;
        INFRA) echo '❌' ;;
        *)     echo '⚪' ;;
    esac
}

release_link() {
    printf '[%s](https://github.com/interuss/dss/releases/tag/interuss%%2Fdss%%2F%s)' "$1" "$1"
}

md_table() {
    local a b cell footnote=0
    printf 'The following matrix shows what is possible when a user wants to upgrade a pool on version A (rows) to\n'
    printf 'version B (columns).\n'
    printf 'The table always assumes a migration to the latest schema of the target version B prior to DSS version upgrade per "Rolling upgrade procedure" below.  Where this cannot be accomplished (e.g., DSS version X cannot function with the latest schema of DSS version X+1), the transition will be indicated as incompatible.\n'
    printf '| A \\ B '
    for b in "${VERSIONS[@]}"; do printf '| %s ' "$(release_link "$b")"; done
    printf '|\n|---'
    for _ in "${VERSIONS[@]}"; do printf '|---'; done
    printf '|\n'
    for a in "${VERSIONS[@]}"; do
        printf '| **%s** ' "$(release_link "$a")"
        for b in "${VERSIONS[@]}"; do
            if (( ${VIDX[$b]} < ${VIDX[$a]} )); then
                cell='⚪'
            elif [[ "$a" == "$b" ]]; then
                cell='✅'
            else
                cell="$(emoji "${VERDICT["$a|$b"]:-}")"
                if [[ "$a" == "v0.20.2" || "$b" == "v0.20.2" ]]; then
                    cell="$cell<sup>1</sup>"
                    footnote=1
                fi
            fi
            printf '| %s ' "$cell"
        done
        printf '|\n'
    done
    printf '✅ compatible · ⚠️ degraded, see explanation below · ❌ incompatible · ⚪ not evaluated\n'
    if (( footnote )); then
        printf '<sup>1</sup>Some tests in a multi-version pool may fail due to improvements in the test suite and DSS behavior.\n'
    fi
}

section "Build dummy-oauth image (if missing)"
if docker image inspect "$OAUTH_IMAGE" >/dev/null 2>&1; then
    ok "$OAUTH_IMAGE already present"
else
    ( cd "$REPO_ROOT" && docker build -f cmds/dummy-oauth/Dockerfile -t "$OAUTH_IMAGE" . ) \
        >"$LOG_DIR/dummy-oauth-build.log" 2>&1
    ok "built $OAUTH_IMAGE"
fi

section "Pull images"
for v in "${VERSIONS[@]}"; do
    docker pull -q "$DSS_REGISTRY:$v" >/dev/null
    info "$DSS_REGISTRY:$v"
done
docker pull -q "$MONITORING_IMAGE" >/dev/null
docker pull -q "$V0_20_2_MONITORING_IMAGE" >/dev/null
docker pull -q "$CRDB_IMAGE" >/dev/null

TOTAL=$(( ${#VERSIONS[@]} * (${#VERSIONS[@]} - 1) / 2 ))
N=0
START=$(date +%s)
for a in "${VERSIONS[@]}"; do
    for b in "${VERSIONS[@]}"; do
        (( ${VIDX[$b]} <= ${VIDX[$a]} )) && continue
        N=$((N + 1))
        section "[$N/$TOTAL] migrations $a  ·  A=$a  B=$b"
        t0=$(date +%s)
        run_combo "$a" "$b"
        printf '  %s  %s  %s(%s)%s\n' "$(symbol "${VERDICT["$a|$b"]}")" "${VERDICT["$a|$b"]}" \
            "$DIM" "${DETAIL["$a|$b"]}" "$RESET"
        info "took $(fmt_dur "$(( $(date +%s) - t0 ))")"
    done
done
teardown

section "Matrix (rows = A; columns = B, migrations applied by B)"
printf '  %-14s' ''
for b in "${VERSIONS[@]}"; do printf '%-13s' "$b"; done
printf '\n'
for a in "${VERSIONS[@]}"; do
    printf '  %-14s' "$a"
    for b in "${VERSIONS[@]}"; do
        if (( ${VIDX[$b]} < ${VIDX[$a]} )); then
            printf '%s·%s' "$DIM" "$RESET"
        elif [[ "$a" == "$b" ]]; then
            printf '%s✓%s' "$GREEN" "$RESET"
        else
            printf '%s' "$(symbol "${VERDICT["$a|$b"]:-}")"
        fi
        printf '%-12s' ''
    done
    printf '\n'
done
printf '\n  %s✓%s pass   %s✗%s fail   %s⨯%s stack did not start\n' \
    "$GREEN" "$RESET" "$RED" "$RESET" "$RED" "$RESET"
printf '  %stotal: %s   logs: %s   reports: %s%s\n' \
    "$DIM" "$(fmt_dur "$(( $(date +%s) - START ))")" "$LOG_DIR" "$QUAL_OUT" "$RESET"

section "Markdown"
echo
md_table
