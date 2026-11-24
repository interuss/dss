#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=release/scripts/common.sh
source "$SCRIPT_DIR/common.sh"

USSQ_OUT="$RELEASE_DIR/uss_qualifier_output"
BENCH_OUT="$RELEASE_DIR/tests/benchmarker/output"
OUTPUT_DIR="$RELEASE_DIR/output"

# subdir  short_name  display
DATASTORES=(
    "cockroachdb crdb Cockroachdb"
    "yugabyte    ybdb Yugabyte"
    "raft        raft Raft"
)

# Extract "v0.22.0" from: image = "docker.io/interuss/dss:v0.22.0"
get_release_version_from_vars() {
    grep -E '^[[:space:]]*image[[:space:]]*=' "$1" 2>/dev/null | head -1 \
        | sed -E 's/.*"[^"]*:([^"]+)".*/\1/'
}

# Extract "v0.31.0" from "codebase_version": "Interuss/monitoring/v0.31.0"
get_monitoring_version_from_report() {
    [[ -f "$1" ]] || { echo "unknown"; return; }
    grep -oE '"codebase_version"[[:space:]]*:[[:space:]]*"[^"]+"' "$1" | head -1 \
        | sed -E 's/.*"([^"]+)"$/\1/' | awk -F/ '{print $NF}'
}

# Redact unrelevent values in .tfvars content.
redact_vars() {
    sed -E \
        -e 's/(aws_route53_zone_id[[:space:]]*=[[:space:]]*)"[^"]*"/\1"[REDACTED]"/' \
        -e 's/(google_project_name[[:space:]]*=[[:space:]]*)"[^"]*"/\1"[REDACTED]"/' \
        -e 's/(google_dns_managed_zone_name[[:space:]]*=[[:space:]]*)"[^"]*"/\1"[REDACTED]"/' \
        "$1"
}

# Personal tfvars if clusters were spawned from this machine, release tfvars otherwise.
tfvars_of() {
    local personal="$PERSONAL_DIR/$1/terraform.tfvars"
    if [[ -f "$personal" ]]; then
        echo "$personal"
    else
        sed "s|\\\$IMAGE|${IMAGE:-}|g" "$RELEASE_INFRA_DIR/$1.tfvars" > "$STAGE/$1.tfvars"
        echo "$STAGE/$1.tfvars"
    fi
}

# === 1. Resolve release version ===
if [[ -n "${IMAGE:-}" ]]; then
    VERSION="${IMAGE##*:}"
else
    for v in "$PERSONAL_DIR"/release-*-dss-*/terraform.tfvars; do
    [[ -f "$v" ]] && VERSION=$(get_release_version_from_vars "$v")
    [[ -n "${VERSION:-}" ]] && break
done
fi
# shellcheck disable=SC2016
[[ -n "${VERSION:-}" ]] || die 'cannot determine release version (set $IMAGE or run spawn-clusters first)'
printf 'release version: %s\n\n' "$VERSION"

mkdir -p "$OUTPUT_DIR"
STAGE="$OUTPUT_DIR/.stage"
rm -rf "$STAGE"
mkdir -p "$STAGE/$VERSION"

# === 2. Build per-datastore directories ===
for entry in "${DATASTORES[@]}"; do
    read -r subdir short display <<< "$entry"
    ds_dir="$STAGE/$VERSION/$subdir"
    mkdir -p "$ds_dir/prober"

    aws_vars=$(tfvars_of "release-aws-dss-$short")
    goo_vars=$(tfvars_of "release-google-dss-$short")
    aws_log="$LOG_DIR/release-aws-dss-$short.prober.log"
    goo_log="$LOG_DIR/release-google-dss-$short.prober.log"
    ussq_src="$USSQ_OUT/$short"
    bench_log="$LOG_DIR/$short.benchmarker.log"
    monitoring_version=$(get_monitoring_version_from_report "$ussq_src/$short/report.json")

    printf '%s» %s%s  (%s)\n' "$BOLD$CYAN" "$subdir" "$RESET" "$display"
    for f in "$aws_vars" "$goo_vars" "$aws_log" "$goo_log" "$bench_log"; do
        [[ -e "$f" ]] || warn "missing: $f"
    done
    [[ -d "$ussq_src" ]] || warn "missing: $ussq_src"

    aws_endpoint=$(get_app_hostname "$aws_vars" || echo "<unknown>")
    goo_endpoint=$(get_app_hostname "$goo_vars" || echo "<unknown>")

    # --- README.md ---
    {
        printf '# Validation of %s with %s datastore\n\n' "$VERSION" "$display"
        cat <<'EOF'
## Infrastructure

The test were performed on a DSS Pool composed of two DSS instances deployed in two distinct cloud providers.
(AWS and Google Cloud) with production configuration (compute and storage).

For reference, the configuration of the deployment used are the ones below.

### AWS - terraform.tfvars

```hcl2
EOF
        [[ -f "$aws_vars" ]] && redact_vars "$aws_vars"
        cat <<'EOF'
```

### Google - terraform.tfvars

```hcl2
EOF
        [[ -f "$goo_vars" ]] && redact_vars "$goo_vars"
        cat <<'EOF'
```

## Services

- The AWS instance was deployed using tanka
- The Google instance was deployed using helm


## Test driver

Prober, USS qualifier and benchmarker test runs have been performed from an AWS EC2 instance (m5.2xlarge, us-east-2) connected to public Internet outside of the Kubernetes clusters hosting the DSS instances.

EOF
        # shellcheck disable=SC2016
        printf 'Monitoring image: `%s`\n\n' "$monitoring_version"
    } > "$ds_dir/README.md"
    printf '  + README.md\n'

    # --- prober txt files ---
    for side in aws google; do
        case $side in
            aws)    src="$aws_log"; dst="$ds_dir/prober/dss1-aws.txt";    ep="$aws_endpoint" ;;
            google) src="$goo_log"; dst="$ds_dir/prober/dss2-google.txt"; ep="$goo_endpoint" ;;
        esac
        {
            printf 'image: interuss/monitoring:%s\n' "$monitoring_version"
            printf 'dss-endpoint: https://%s\n' "$ep"
            [[ -f "$src" ]] && cat "$src"
        } > "$dst"
        printf '  + prober/%s\n' "$(basename "$dst")"
    done

    # --- uss_qualifier output ---
    ussq_dst="$ds_dir/uss_qualifier"
    if [[ -d "$ussq_src" ]]; then
        cp -R "$ussq_src" "$ussq_dst"
        n=$(find "$ussq_dst" -type f | wc -l)
        printf '  + uss_qualifier/  (%d files)\n' "$n"
    fi

    # --- benchmarker output ---
    bench_dst="$ds_dir/benchmarker"
    mkdir -p "$bench_dst"
    for src in "$BENCH_OUT"/*_"$short"; do
        [[ -d "$src" ]] || { warn "missing: $BENCH_OUT/*_$short"; continue; }
        test_name=$(basename "$src" "_$short")
        mkdir -p "$bench_dst/$test_name"
        find "$src" -maxdepth 1 -name '*.png' -exec cp {} "$bench_dst/$test_name/" \;
        n=$(find "$bench_dst/$test_name" -type f | wc -l)
        printf '  + benchmarker/%s/  (%d files)\n' "$test_name" "$n"
    done
done

# === 3. Zip ===
ZIP="$OUTPUT_DIR/$VERSION.zip"
rm -f "$ZIP"
( cd "$STAGE" && zip -rq "$ZIP" "$VERSION" )
rm -rf "$STAGE"

section "Done"
printf '  %s%s%s  (%s)\n\n' "$BOLD$GREEN" "$ZIP" "$RESET" "$(du -h "$ZIP" | cut -f1)"
unzip -l "$ZIP" | tail -8
