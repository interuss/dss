#!/usr/bin/env bash

set -eo pipefail

# Find and change to repo root directory
OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi
cd "${BASEDIR}/.." || exit 1

docker build \
    -f monitoring/Dockerfile \
    -t interuss/monitoring \
    --build-arg version="$(scripts/git/version.sh monitoring --long)" \
    monitoring \
  || exit 1
