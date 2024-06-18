#!/usr/bin/env bash

# This script builds InterUSS docker images and may be run from any working
# directory.  If run without a DOCKER_URL environment variable, it will just
# build images named interuss-local/*.  If DOCKER_URL is present, it will both
# build the versioned dss image and push it to the DOCKER_URL remote.
# If DOCKER_URL is set, DOCKER_UPDATE_LATEST can be optionally set to `true` in order
# to publish the latest tag along the version.

set -eo pipefail

OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")/.."
else
	BASEDIR=$(readlink -e "$(dirname "$0")/..")
fi
cd "${BASEDIR}"

VERSION=$(./scripts/git/version.sh dss)
LATEST_TAG="latest"

if [[ -z "${DOCKER_URL}" ]]; then
  echo "DOCKER_URL environment variable is not set; building image to interuss-local/dss..."
  docker image build . -t interuss-local/dss

  echo "DOCKER_URL environment variable was not set; built images to interuss-local/dss"
else
  echo "Building image ${DOCKER_URL}/dss:${VERSION}"
  docker image build . -t "${DOCKER_URL}/dss:${VERSION}"

  echo "Pushing docker image ${DOCKER_URL}/dss:${VERSION}..."
  docker image push "${DOCKER_URL}/dss:${VERSION}"

  echo "Built and pushed docker image ${DOCKER_URL}/dss:${VERSION}"

  if [[ "${DOCKER_UPDATE_LATEST}" == "true" ]]; then

    echo "Tagging docker image ${DOCKER_URL}/dss:${LATEST_TAG}..."
    docker tag "${DOCKER_URL}/dss:${VERSION}" "${DOCKER_URL}/dss:${LATEST_TAG}"

    echo "Pushing docker image ${DOCKER_URL}/dss:${LATEST_TAG}..."

    docker image push "${DOCKER_URL}/dss:${LATEST_TAG}"

    echo "Built and pushed docker image ${DOCKER_URL}/dss:${LATEST_TAG}"

  fi

  echo "VAR_DOCKER_IMAGE_NAME: ${DOCKER_URL}/dss:${VERSION}"
fi
