#!/usr/bin/env bash
set -e

set -eo pipefail

OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")/.."
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi
cd "${BASEDIR}"

docker build -t terraform-variables:latest .
docker run \
    -v "${SCRIPT_DIR}/":/app/utils:rw \
    -v "${SCRIPT_DIR}/../dependencies":/app/examples:rw \
    -v "${SCRIPT_DIR}/../modules":/app/modules:rw \
    -w /app/utils \
    terraform-variables \
    ls ../ && python variables.py $@
