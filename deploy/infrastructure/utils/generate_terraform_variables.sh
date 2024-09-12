#!/usr/bin/env bash
set -e

set -eo pipefail

OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi
cd "${BASEDIR}"

docker build -t terraform-variables:latest .
docker run \
    -v "${BASEDIR}/":/app/infrastructure/utils:rw \
    -v "${BASEDIR}/../dependencies":/app/infrastructure/dependencies:rw \
    -v "${BASEDIR}/../modules":/app/infrastructure/modules:rw \
    -v "${BASEDIR}/../../operations":/app/operations:rw \
    -w /app/infrastructure/utils/ \
    terraform-variables \
    /bin/bash -c "python variables.py $*"
