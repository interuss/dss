#!/usr/bin/env bash

# shellcheck disable=SC2034

# Shared config + helpers for release scripts.
# Source from a script that has SCRIPT_DIR set.

RELEASE_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
PERSONAL_DIR="$REPO_ROOT/deploy/infrastructure/personal"
RELEASE_INFRA_DIR="$RELEASE_DIR/infrastructure"
LOG_DIR="$RELEASE_DIR/logs"

HELM_RELEASE="dss"
TANKA_WORKSPACES_ROOT="$REPO_ROOT/deploy/services/tanka/workspace"

# Cluster list derived from infrastructure/*.tfvars filenames.
mapfile -t CLUSTERS < <(
    find "$RELEASE_INFRA_DIR" -maxdepth 1 -type f -name '*.tfvars' -printf '%f\n' 2>/dev/null \
        | sed -E 's/\.tfvars$//' | sort
)
(( ${#CLUSTERS[@]} > 0 )) || { printf 'ERROR: no *.tfvars found in %s\n' "$RELEASE_INFRA_DIR" >&2; exit 1; }

if [[ -t 1 ]]; then
    BOLD=$'\e[1m'; DIM=$'\e[2m'; RESET=$'\e[0m'
    RED=$'\e[31m'; GREEN=$'\e[32m'; YELLOW=$'\e[33m'; BLUE=$'\e[34m'; CYAN=$'\e[36m'
else
    BOLD=""; DIM=""; RESET=""; RED=""; GREEN=""; YELLOW=""; BLUE=""; CYAN=""
fi

hr()       { printf '%s\n' "${DIM}────────────────────────────────────────────────────────────${RESET}"; }
section()  { printf '\n%s» %s%s\n' "$BOLD$CYAN" "$1" "$RESET"; hr; }
fmt_dur()  { local s=$1; printf '%dm%02ds' "$((s/60))" "$((s%60))"; }

ok()       { printf '  %s✓%s  %s\n' "$GREEN" "$RESET" "$*"; }
fail()     { printf '  %s✗%s  %s\n' "$RED"   "$RESET" "$*"; }
info()     { printf '  %s•%s  %s\n' "$DIM"   "$RESET" "$*"; }
start()    { printf '  %s▶%s  %s\n' "$BLUE"  "$RESET" "$*"; }
warn()     { printf '  %s!%s  %s\n' "$YELLOW" "$RESET" "$*"; }
die()      { printf '%sERROR%s %s\n' "$RED" "$RESET" "$*" >&2; exit 1; }

# Return the cloud component of a cluster name.
cloud_of() {
    case "$1" in
        *aws*)    echo aws ;;
        *google*) echo google ;;
        *)        return 1 ;;
    esac
}

# Extract app_hostname from a .tfvars file.
get_app_hostname() {
    grep -E '^[[:space:]]*app_hostname[[:space:]]*=' "$1" 2>/dev/null | head -1 \
        | sed -E 's/^[[:space:]]*app_hostname[[:space:]]*=[[:space:]]*"([^"]+)".*/\1/'
}

# Parallel job tracker. Caller registers PID_OF[$name] and START_OF[$name],
# then calls wait_jobs <log_suffix> <name...>. Prints per-job result,
# emits a heartbeat every 30s, returns 1 on any failure, and clears state.
declare -gA PID_OF=() START_OF=()
wait_jobs() {
    local suffix="$1"; shift
    local -a names=("$@")
    local total=${#names[@]}
    local failed=0 done_count=0
    local batch_start last_heartbeat now dur running name
    batch_start=$(date +%s)
    last_heartbeat=$batch_start
    declare -A done_local=()
    while (( done_count < total )); do
        for name in "${names[@]}"; do
            [[ -n "${done_local[$name]:-}" ]] && continue
            if ! kill -0 "${PID_OF[$name]}" 2>/dev/null; then
                dur=$(( $(date +%s) - START_OF[$name] ))
                if wait "${PID_OF[$name]}"; then
                    printf '  %s✓%s  %-32s done   in %s\n' \
                        "$GREEN" "$RESET" "$name" "$(fmt_dur "$dur")"
                else
                    printf '  %s✗%s  %-32s FAILED in %s  → %s\n' \
                        "$RED" "$RESET" "$name" "$(fmt_dur "$dur")" \
                        "$LOG_DIR/$name.$suffix.log"
                    failed=1
                fi
                done_local[$name]=1
                done_count=$((done_count + 1))
            fi
        done
        if (( done_count < total )); then
            sleep 5
            now=$(date +%s)
            if (( now - last_heartbeat >= 30 )); then
                running=""
                for name in "${names[@]}"; do
                    [[ -z "${done_local[$name]:-}" ]] && running+="$name "
                done
                printf '  %s⏳%s  %s%s%s  elapsed %s - still running: %s\n' \
                    "$YELLOW" "$RESET" "$DIM" "$(date +%H:%M:%S)" "$RESET" \
                    "$(fmt_dur "$((now - batch_start))")" "${running% }"
                last_heartbeat=$now
            fi
        fi
    done
    PID_OF=()
    START_OF=()
    return $failed
}
