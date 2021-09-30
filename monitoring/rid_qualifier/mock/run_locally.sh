#!/usr/bin/env bash

PORT=8070

OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi

# Always start from the repo folder root
cd "${BASEDIR}/../../.." || exit 1

echo Building rid_qualifier mock...
docker build \
    -f monitoring/rid_qualifier/mock/Dockerfile \
    -t interuss/automated-testing/rid-qualifier/mock \
    --build-arg version="$(scripts/git/commit.sh)" \
    monitoring \
    || exit 1

echo Running rid_qualifier mock...
docker run --name rid_qualifier_mock \
  --rm \
  -p ${PORT}:5000 \
  interuss/automated-testing/rid-qualifier/mock
