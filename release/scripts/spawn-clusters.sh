#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=release/scripts/common.sh
source "$SCRIPT_DIR/common.sh"

MODULES_DIR="$REPO_ROOT/deploy/infrastructure/modules"
MODULE_FILES=(main.tf output.tf variables.gen.tf)

[[ -n "${IMAGE:-}" ]] || die 'IMAGE env var is required (e.g. export IMAGE=docker.io/interuss/dss:v0.22.0)'
[[ -n "${ZONE_ID:-}" ]] || die 'ZONE_ID env var is required (e.g. export ZONE_ID=0AWSZONEID0)'
[[ -n "${GOOGLE_PROJECT_NAME:-}" ]] || die 'GOOGLE_PROJECT_NAME env var is required (e.g. export GOOGLE_PROJECT_NAME=interuss)'
[[ -n "${ZONE_NAME:-}" ]] || die 'ZONE_NAME env var is required (e.g. export ZONE_NAME=zonename)'
command -v envsubst >/dev/null || die "envsubst not installed (install gettext-base)"

rm -f "$LOG_DIR"/*.spawn.log 2>/dev/null || true

module_for() {
    local cloud
    cloud=$(cloud_of "$1") || die "cannot infer cloud for $1"
    echo "terraform-$cloud-dss"
}

# === 1. Cleanup ===
section "Cleanup personal configs"
for name in "${CLUSTERS[@]}"; do
    if [[ -d "$PERSONAL_DIR/$name" ]]; then
        printf '  %s✗%s  rm  %s\n' "$YELLOW" "$RESET" "$PERSONAL_DIR/$name"
        rm -rf "${PERSONAL_DIR:?}/${name:?}"
    else
        printf '  %s·%s  ok  %s (absent)\n' "$DIM" "$RESET" "$PERSONAL_DIR/$name"
    fi
done

# === 2. Copy module + release tfvars into personal/ ===
section "Copy configs into personal/"
for name in "${CLUSTERS[@]}"; do
    module=$(module_for "$name")
    src_module="$MODULES_DIR/$module"
    src_release="$RELEASE_INFRA_DIR/$name.tfvars"
    dst="$PERSONAL_DIR/$name"

    [[ -d "$src_module"  ]] || die "missing module $src_module"
    [[ -f "$src_release" ]] || die "missing release tfvars $src_release"

    mkdir -p "$dst"
    for f in "${MODULE_FILES[@]}"; do
        [[ -f "$src_module/$f" ]] || die "$f missing in $src_module"
        cp "$src_module/$f" "$dst/"
    done

    # shellcheck disable=SC2016
    envsubst '$IMAGE,$ZONE_ID,$GOOGLE_PROJECT_NAME,$ZONE_NAME' < "$src_release" > "$dst/terraform.tfvars"

    printf '  %s+%s  %s%s%s  ← module:%s + %s.tfvars\n' \
        "$GREEN" "$RESET" "$BOLD" "$name" "$RESET" "$module" "$name"
done

# === 3. Spawn in parallel ===
section "Spawn ${#CLUSTERS[@]} clusters in parallel"
printf '  %slogs:%s %s\n\n' "$DIM" "$RESET" "$LOG_DIR"

BATCH_START=$(date +%s)
for name in "${CLUSTERS[@]}"; do
    START_OF[$name]=$(date +%s)
    (
        cd "$PERSONAL_DIR/$name"
        terraform init -input=false && terraform apply -auto-approve -input=false
    ) >"$LOG_DIR/$name.spawn.log" 2>&1 &
    PID_OF[$name]=$!
    printf '  %s▶%s  %-32s started (pid %d)\n' "$BLUE" "$RESET" "$name" "${PID_OF[$name]}"
done
echo
info "Waiting (heartbeat every 30s)..."
echo

failed=0
wait_jobs spawn "${CLUSTERS[@]}" || failed=1

# === 4. Summary ===
TOTAL_DUR=$(( $(date +%s) - BATCH_START ))
section "Summary"
if (( failed == 0 )); then
    printf '  %sAll %d clusters up%s in %s\n' \
        "$BOLD$GREEN" "${#CLUSTERS[@]}" "$RESET" "$(fmt_dur "$TOTAL_DUR")"
else
    printf '  %sSome clusters failed%s - total time %s\n' \
        "$BOLD$RED" "$RESET" "$(fmt_dur "$TOTAL_DUR")"
fi

exit "$failed"
