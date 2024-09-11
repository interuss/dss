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
    -v "${BASEDIR}/":/app/utils:rw \
    -v "${BASEDIR}/../dependencies":/app/dependencies:rw \
    -v "${BASEDIR}/../modules":/app/modules:rw \
    -v "${BASEDIR}/../../operations":/operations:rw \
    -w /app/utils \
    terraform-variables \
    /bin/bash -c "python variables.py $*"
