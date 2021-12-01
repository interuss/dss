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
cd "${BASEDIR}/../../../.." || exit 1

DOCKER_FILE="monitoring/uss_qualifier/rid/mock/Dockerfile"
DOCKER_TAG="interuss/uss-qualifier/rid-mock"

echo Building RID mock...
docker build \
    -f "${DOCKER_FILE}" \
    -t "${DOCKER_TAG}" \
    --build-arg version="$(scripts/git/commit.sh)" \
    monitoring \
    || exit 1

echo Running RID mock...
docker run --name uss_qualifier_rid_mock \
  --rm \
  -p ${PORT}:5000 \
  "${DOCKER_TAG}"
