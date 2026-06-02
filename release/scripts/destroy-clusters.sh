#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=release/scripts/common.sh
source "$SCRIPT_DIR/common.sh"

rm -f "$LOG_DIR"/*.destroy.log 2>/dev/null || true

# === 1. Build target list ===
section "Destroy clusters in parallel"
printf '  %slogs:%s %s\n\n' "$DIM" "$RESET" "$LOG_DIR"

TARGETS=()
for name in "${CLUSTERS[@]}"; do
    if [[ ! -d "$PERSONAL_DIR/$name" ]]; then
        printf '  %s·%s  %-32s skipped (no personal config)\n' "$DIM" "$RESET" "$name"
        continue
    fi
    TARGETS+=("$name")
done

if (( ${#TARGETS[@]} == 0 )); then
    printf '\n%sNothing to destroy.%s\n' "$YELLOW" "$RESET"
    exit 0
fi

# === 2. Confirm + countdown ===
printf '\n%sAbout to destroy:%s\n' "$BOLD$RED" "$RESET"
for name in "${TARGETS[@]}"; do
    printf '  %s-%s  %s\n' "$RED" "$RESET" "$name"
done
read -r -p "$(printf '\n%sProceed? [y/N]:%s ' "$BOLD" "$RESET")" ans
if [[ ! "$ans" =~ ^[yY]([eE][sS])?$ ]]; then
    printf '%sAborted.%s\n' "$YELLOW" "$RESET"
    exit 0
fi
for i in 5 4 3 2 1; do
    printf '\r  %s⏳ Destroying in %d... (Ctrl-C to cancel)%s   ' "$YELLOW" "$i" "$RESET"
    sleep 1
done
printf '\r%-60s\n' " "

# === 3. Pre-destroy cleanup: uninstall services, delete PVCs (best-effort) ===
section "Uninstall services + delete PVCs"
declare -A CTX=() TANKA_WS=()
for name in "${TARGETS[@]}"; do
    ctx=$(terraform -chdir="$PERSONAL_DIR/$name" output -raw cluster_context 2>/dev/null || true)
    CTX[$name]="$ctx"
    if [[ "$(cloud_of "$name" || true)" == aws ]]; then
        ws=$(terraform -chdir="$PERSONAL_DIR/$name" output -raw workspace_location 2>/dev/null || true)
        [[ -n "$ws" ]] && TANKA_WS[$name]="$TANKA_WORKSPACES_ROOT/$(basename "$ws")"
    fi
done

for name in "${TARGETS[@]}"; do
    printf '\n  %s── %s%s\n' "$BOLD$CYAN" "$name" "$RESET"
    log="$LOG_DIR/$name.uninstall.log"
    : > "$log"

    case "$(cloud_of "$name" || true)" in
        aws)
            if [[ -d "${TANKA_WS[$name]:-}" ]]; then
                if ( cd "${TANKA_WS[$name]}" && tk delete --auto-approve always . ) >>"$log" 2>&1; then
                    ok "tk delete"
                else
                    warn "tk delete failed (continuing) - see $log"
                fi
            else
                warn "no tanka workspace - skipping tk delete"
            fi
            ;;
        google)
            if helm uninstall "$HELM_RELEASE" --kube-context="${CTX[$name]}" >>"$log" 2>&1; then
                ok "helm uninstall $HELM_RELEASE"
            else
                warn "helm uninstall failed (continuing) - see $log"
            fi
            ;;
    esac

    if [[ -n "${CTX[$name]}" ]]; then
        if kubectl --context="${CTX[$name]}" --request-timeout=120s \
                delete pvc --all -A --wait=true >>"$log" 2>&1; then
            ok "deleted all PVCs"
        else
            warn "delete PVCs failed (continuing) - see $log"
        fi
    else
        warn "no context - skipping PVC delete"
    fi
done

# === 4. Destroy in parallel ===
section "Terraform destroy"
BATCH_START=$(date +%s)
for name in "${TARGETS[@]}"; do
    START_OF[$name]=$(date +%s)
    (
        cd "$PERSONAL_DIR/$name"
        terraform destroy -auto-approve -input=false
    ) >"$LOG_DIR/$name.destroy.log" 2>&1 &
    PID_OF[$name]=$!
    printf '  %s▶%s  %-32s destroying (pid %d)\n' "$BLUE" "$RESET" "$name" "${PID_OF[$name]}"
done
echo
info "Waiting (heartbeat every 30s)..."
echo

failed=0
wait_jobs destroy "${TARGETS[@]}" || failed=1

# === 4. Summary ===
TOTAL_DUR=$(( $(date +%s) - BATCH_START ))
section "Summary"
if (( failed == 0 )); then
    printf '  %sAll %d clusters destroyed%s in %s\n' \
        "$BOLD$GREEN" "${#TARGETS[@]}" "$RESET" "$(fmt_dur "$TOTAL_DUR")"
else
    printf '  %sSome destroys failed%s - total time %s\n' \
        "$BOLD$RED" "$RESET" "$(fmt_dur "$TOTAL_DUR")"
fi

exit "$failed"
