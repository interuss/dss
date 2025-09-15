#!/bin/bash

set -e 

# shellcheck source=dest_paths.var
source dest_paths.var 

for DEST in "${DEST_PATHS[@]}"; do
    DEST_PATH="$DEST/db_versions"

    echo "Cloning to $DEST_PATH."

    mkdir -p "$DEST_PATH"

    cp -r crdb "$DEST_PATH/"
    cp -r yugabyte "$DEST_PATH/"

    echo "Cloned to $DEST_PATH successfully."
done 