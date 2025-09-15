#!/bin/bash

set -e 

SCRIPT_NAME="$(basename "$0")"

DEST_PATHS=(
"../../../deploy/infrastructure/dependencies/terraform-commons-dss"
"../../../deploy/services/helm-charts/dss"
"../../../deploy/services/tanka"
"../../../pkg"
)

for DEST in "${DEST_PATHS[@]}"; do
    DEST_PATH="$DEST/db_versions"

    echo "Cloning to $DEST_PATH."

    mkdir -p "$DEST_PATH"

    cp -r . "$DEST_PATH/"

    rm -f "$DEST_PATH/$SCRIPT_NAME"

    echo "Cloned to $DEST_PATH successfully."
done 