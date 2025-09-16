#!/bin/bash

set -e 

# shellcheck source=./build/db_schemas/version/dest_paths.var
source dest_paths.var 

for DEST in "${DEST_PATHS[@]}"; do
    DEST_PATH="$DEST/db_versions"

    echo "Cloning to $DEST_PATH."

    mkdir -p "$DEST_PATH"

    rm -r "$DEST_PATH/crdb"
    rm -r "$DEST_PATH/yugabyte"

    cp -r crdb "$DEST_PATH/"
    cp -r yugabyte "$DEST_PATH/"

    echo "Cloned to $DEST_PATH successfully."
done 