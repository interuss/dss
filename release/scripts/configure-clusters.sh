#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=release/scripts/common.sh
source "$SCRIPT_DIR/common.sh"

CRDB_CERT_DIRS=(ca_certs_dir node_certs_dir client_certs_dir prometheus_certs_dir)

# === 1. Resolve workspace dirs from terraform output ===
section "Resolving workspaces"
declare -A WS=()
for name in "${CLUSTERS[@]}"; do
    [[ -d "$PERSONAL_DIR/$name" ]] \
        || die "missing personal config $PERSONAL_DIR/$name - run spawn-clusters.sh first"
    ws=$(terraform -chdir="$PERSONAL_DIR/$name" output -raw workspace_location)
    WS[$name]="$ws"
    printf '  %s•%s  %-32s → %s\n' "$DIM" "$RESET" "$name" "$ws"
done

# === 2. Get credentials ===
section "Getting credentials"
for name in "${CLUSTERS[@]}"; do
    printf '  %s→%s  %-32s ' "$BLUE" "$RESET" "$name"
    if ( cd "${WS[$name]}" && ./get-credentials.sh ) >"$LOG_DIR/$name.creds.log" 2>&1; then
        printf '%s✓%s\n' "$GREEN" "$RESET"
    else
        printf '%s✗%s see %s\n' "$RED" "$RESET" "$LOG_DIR/$name.creds.log"
        exit 1
    fi
done

# === 3. Group clusters by datastore type and apply per-type cert ritual ===
# Cluster names follow: release-{cloud}-dss-{datastore}
declare -A CLUSTER_AWS=() CLUSTER_GOOGLE=() ALL_DATASTORES=()
for name in "${CLUSTERS[@]}"; do
    cloud=$(cloud_of "$name") || { warn "$name: unknown cloud, skipping"; continue; }
    ds="${name##*-}"
    ALL_DATASTORES[$ds]=1
    case "$cloud" in
        aws)    CLUSTER_AWS[$ds]="$name" ;;
        google) CLUSTER_GOOGLE[$ds]="$name" ;;
    esac
done

mapfile -t datastores < <(printf '%s\n' "${!ALL_DATASTORES[@]}" | sort)
for ds in "${datastores[@]}"; do
    aws_name="${CLUSTER_AWS[$ds]:-}"
    goo_name="${CLUSTER_GOOGLE[$ds]:-}"
    if [[ -z "$aws_name" || -z "$goo_name" ]]; then
        section "$ds certs"
        info "skipping (incomplete pair: aws=${aws_name:-—}, google=${goo_name:-—})"
        continue
    fi
    aws_ws="${WS[$aws_name]}"
    goo_ws="${WS[$goo_name]}"
    aws_log="$LOG_DIR/$aws_name.certs.log"
    goo_log="$LOG_DIR/$goo_name.certs.log"

    case "$ds" in
        ybdb)
            section "Yugabyte certs ($aws_name ↔ $goo_name)"
            ( cd "$aws_ws" && ./dss-certs.sh init ) >"$aws_log" 2>&1
            ok "init $aws_name"
            ( cd "$goo_ws" && ./dss-certs.sh init ) >"$goo_log" 2>&1
            ok "init $goo_name"

            ( cd "$goo_ws" && ./dss-certs.sh get-ca ) \
                | ( cd "$aws_ws" && ./dss-certs.sh add-pool-ca ) >>"$aws_log" 2>&1
            ok "$aws_name trusts $goo_name"

            ( cd "$aws_ws" && ./dss-certs.sh get-ca ) \
                | ( cd "$goo_ws" && ./dss-certs.sh add-pool-ca ) >>"$goo_log" 2>&1
            ok "$goo_name trusts $aws_name"

            ( cd "$aws_ws" && ./dss-certs.sh apply ) >>"$aws_log" 2>&1
            ok "apply $aws_name"
            ( cd "$goo_ws" && ./dss-certs.sh apply ) >>"$goo_log" 2>&1
            ok "apply $goo_name"
            ;;
        crdb)
            section "CockroachDB certs ($aws_name ↔ $goo_name)"
            ( cd "$aws_ws" && ./make-certs.sh ) >"$aws_log" 2>&1
            ok "make-certs $aws_name"
            ( cd "$goo_ws" && ./make-certs.sh "$aws_ws/ca_certs_dir/ca.crt" ) >"$goo_log" 2>&1
            ok "make-certs $goo_name"

            # Snapshot fresh per-cluster CAs before cross-cat, else we'd append into the source.
            aws_ca=$(mktemp); goo_ca=$(mktemp)
            cp "$aws_ws/ca_certs_dir/ca.crt" "$aws_ca"
            cp "$goo_ws/ca_certs_dir/ca.crt" "$goo_ca"

            # apply-certs.sh creates k8s secrets from these dirs, so patch ca.crt in all
            # of them so the cockroachdb.{ca.crt,node,client.root,...} secrets carry both CAs.
            for d in "${CRDB_CERT_DIRS[@]}"; do
                cat "$goo_ca" >> "$aws_ws/$d/ca.crt"
                cat "$aws_ca" >> "$goo_ws/$d/ca.crt"
            done
            rm -f "$aws_ca" "$goo_ca"
            ok "appended foreign CA into ${CRDB_CERT_DIRS[*]} on both sides"

            ( cd "$aws_ws" && ./apply-certs.sh ) >>"$aws_log" 2>&1
            ok "apply-certs $aws_name"
            ( cd "$goo_ws" && ./apply-certs.sh ) >>"$goo_log" 2>&1
            ok "apply-certs $goo_name"
            ;;
        *)
            section "$ds certs"
            warn "unknown datastore type - no cert ritual defined"
            ;;
    esac
done

section "All ${#CLUSTERS[@]} clusters prepared"
info "kubeconfig + certs ready."
