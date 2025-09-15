#!/bin/bash

set -e

source dest_paths.var

# Function to compare directories
compare_dirs() {
    local src_dir="$1"
    local dest_dir="$2"

    if ! diff -qr "$src_dir" "$dest_dir"; then
        echo "Difference found between $src_dir and $dest_dir, please run the clone.sh script in /build/db_schemas/version."
        return 1
    fi
}

echo "Checking that db versions are up to date."

for DEST in "${DEST_PATHS[@]}"; do
    DEST_PATH="$DEST/db_versions"

    compare_dirs "crdb" "$DEST_PATH/crdb"
    compare_dirs "yugabyte" "$DEST_PATH/yugabyte"
done

echo "All db versions are up to date."