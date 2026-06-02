#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=release/scripts/common.sh
source "$SCRIPT_DIR/common.sh"

HELM_CHART="$REPO_ROOT/deploy/services/helm-charts/dss"
HEALTH_POLL_INTERVAL=5
STATUS_REFRESH_INTERVAL=20

rm -f "$LOG_DIR"/*.apply.log "$LOG_DIR"/*.healthy 2>/dev/null || true

# Kill background jobs on exit / Ctrl-C.
cleanup() {
    local pids
    pids=$(jobs -p)
    # shellcheck disable=SC2015,SC2086
    [[ -n "$pids" ]] && kill $pids 2>/dev/null || true
}
trap cleanup EXIT INT TERM

# === 1. Resolve per-cluster metadata ===
section "Resolve workspaces / contexts / hostnames"
declare -A WS=() CTX=() HOST=() TANKA_WS=()
for name in "${CLUSTERS[@]}"; do
    [[ -d "$PERSONAL_DIR/$name" ]] \
        || die "missing $PERSONAL_DIR/$name - run spawn-clusters.sh first"
    ws=$(terraform -chdir="$PERSONAL_DIR/$name" output -raw workspace_location)
    ctx=$(terraform -chdir="$PERSONAL_DIR/$name" output -raw cluster_context)
    host=$(get_app_hostname "$RELEASE_INFRA_DIR/$name.vars")
    [[ -n "$host" ]] || die "app_hostname not found in $name.vars"
    WS[$name]="$ws"
    CTX[$name]="$ctx"
    HOST[$name]="$host"
    TANKA_WS[$name]="$TANKA_WORKSPACES_ROOT/$(basename "$ws")"
    printf '  %s•%s  %-32s ctx=%s host=%s\n' "$DIM" "$RESET" "$name" "$ctx" "$host"
done

# === 2. Helm dep update (chart-wide, once) ===
section "helm dep update"
( cd "$HELM_CHART" && helm dep update ) >"$LOG_DIR/helm-dep-update.log" 2>&1
ok "helm dep update"

# === 3. Apply states in parallel (tanka on aws, helm on gke) ===
section "Apply states in parallel"
printf '  %slogs:%s %s\n\n' "$DIM" "$RESET" "$LOG_DIR"

apply_one() {
    local name="$1"
    case "$(cloud_of "$name")" in
        aws)
            cd "${TANKA_WS[$name]}"
            tk apply --auto-approve always .
            ;;
        google)
            cd "$HELM_CHART"
            helm upgrade --install \
                --kube-context="${CTX[$name]}" \
                -f "${WS[$name]}/helm_values.yml" \
                "$HELM_RELEASE" .
            ;;
    esac
}

APPLY_START=$(date +%s)
for name in "${CLUSTERS[@]}"; do
    START_OF[$name]=$(date +%s)
    ( apply_one "$name" ) >"$LOG_DIR/$name.apply.log" 2>&1 &
    PID_OF[$name]=$!
    tool=$([[ "$(cloud_of "$name")" == aws ]] && echo tanka || echo helm)
    printf '  %s▶%s  %-32s started (pid %d, %s)\n' "$BLUE" "$RESET" "$name" "${PID_OF[$name]}" "$tool"
done
echo

apply_failed=0
wait_jobs apply "${CLUSTERS[@]}" || apply_failed=1

if (( apply_failed != 0 )); then
    printf '\n%sApply failed - skipping health-wait%s\n' "$RED" "$RESET"
    exit 1
fi
printf '  %stotal apply time: %s%s\n' "$DIM" "$(fmt_dur "$(( $(date +%s) - APPLY_START ))")" "$RESET"

# === 4. Wait for /healthy on each cluster, show unhealthy pods while waiting ===
section "Wait for https://<DSS>/healthy"

service_healthy() {
    local host="$1" marker="$2"
    while true; do
        if body=$(curl -fsk --max-time 5 "https://$host/healthy" 2>/dev/null) && [[ "$body" == "ok" ]]; then
            touch "$marker"
            return 0
        fi
        sleep "$HEALTH_POLL_INTERVAL"
    done
}

declare -A HEALTHY_MARKER=()
for name in "${CLUSTERS[@]}"; do
    HEALTHY_MARKER[$name]="$LOG_DIR/$name.healthy"
    service_healthy "${HOST[$name]}" "${HEALTHY_MARKER[$name]}" &
done

show_unhealthy_pods() {
    local ctx="$1" name="$2"
    printf '\n  %s── %s%s  %s(ctx=%s)%s\n' "$BOLD$CYAN" "$name" "$RESET" "$DIM" "$ctx" "$RESET"
    kubectl --context="$ctx" get pods -A --no-headers 2>/dev/null \
        | awk -v R="$RED" -v Y="$YELLOW" -v G="$GREEN" -v X="$RESET" '
            BEGIN { n = 0 }
            {
                if ($4 == "Completed") next
                split($3, r, "/")
                if ($4 == "Running" && r[1] == r[2]) next
                col = Y
                if ($4 ~ /Error|Crash|Fail|OOM|Evicted|Unknown|ImagePull|InvalidImage/) col = R
                printf "    %s%s%s\n", col, $0, X
                n++
            }
            END { if (n == 0) printf "    %s(all pods ready)%s\n", G, X }
        '
}

HEALTH_START=$(date +%s)
while true; do
    healthy_count=0
    status_line=""
    for name in "${CLUSTERS[@]}"; do
        if [[ -f "${HEALTHY_MARKER[$name]}" ]]; then
            healthy_count=$((healthy_count + 1))
            status_line+="${GREEN}✓${RESET} $name  "
        else
            status_line+="${YELLOW}⏳${RESET} $name  "
        fi
    done
    elapsed=$(( $(date +%s) - HEALTH_START ))
    printf '\n%s[%s] elapsed %s%s - %s\n' "$BOLD" "$(date +%H:%M:%S)" "$(fmt_dur "$elapsed")" "$RESET" "$status_line"

    (( healthy_count == ${#CLUSTERS[@]} )) && break

    for name in "${CLUSTERS[@]}"; do
        [[ -f "${HEALTHY_MARKER[$name]}" ]] && continue
        show_unhealthy_pods "${CTX[$name]}" "$name"
    done
    sleep "$STATUS_REFRESH_INTERVAL"
done

section "All ${#CLUSTERS[@]} clusters healthy"
printf '  total health-wait: %s\n' "$(fmt_dur "$(( $(date +%s) - HEALTH_START ))")"
