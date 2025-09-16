#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

DEST_PATHS=(
"../../../deploy/infrastructure/dependencies/terraform-commons-dss"
"../../../deploy/services/helm-charts/dss"
"../../../deploy/services/tanka"
"../../../pkg"
)

check_versions() {
    echo "Checking that db versions are up to date..."

    for DEST in "${DEST_PATHS[@]}"; do
        DEST_PATH="$DEST/db_versions"

        for DB in crdb yugabyte; do
            if ! diff -qr "$SCRIPT_DIR/$DB" "$DEST_PATH/$DB"; then
                echo "Difference found between $DB and $DEST_PATH/$DB"
                echo "Please run: ./versions.sh clone"
                return 1
            fi
        done
    done

    echo "All db versions are up to date."
}

clone_versions() {
    echo "Cloning db versions"

    for DEST in "${DEST_PATHS[@]}"; do
        DEST_PATH="$DEST/db_versions"

        echo "Cloning to $DEST_PATH"

        mkdir -p "$DEST_PATH"

        rm -rf "$DEST_PATH/crdb"
        rm -rf "$DEST_PATH/yugabyte"

        cp -r "$SCRIPT_DIR/crdb" "$DEST_PATH/"
        cp -r "$SCRIPT_DIR/yugabyte" "$DEST_PATH/"

        echo "Cloned to $DEST_PATH successfully."
    done
}

# Main logic: check args
case "$1" in
    clone)
        clone_versions
        ;;
    check)
        check_versions
        ;;
    *)
        echo "Usage: $0 {clone|check}"
        exit 1
        ;;
esac
