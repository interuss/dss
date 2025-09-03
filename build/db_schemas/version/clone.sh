#!/bin/bash

set -e 

SCRIPT_NAME="$(basename "$0")"

DEST_PATH="../../../deploy/services/helm-charts/dss/version"

mkdir -p "$DEST_PATH"

cp -r . "$DEST_PATH/"

rm -f "$DEST_PATH/$SCRIPT_NAME"

echo "Cloned version folder to $DEST_PATH successfully."
