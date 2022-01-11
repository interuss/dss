#!/usr/bin/env bash

# This script builds InterUSS docker images and may be run from any working
# directory.  If run without a DOCKER_URL environment variable, it will just
# build images named interuss-local/*.  If DOCKER_URL is present, it will both
# build the versioned dss image and push it to the DOCKER_URL remote.

set -eo pipefail

OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")/.."
else
	BASEDIR=$(readlink -e "$(dirname "$0")/..")
fi
cd "${BASEDIR}"

VERSION=$(./scripts/git/version.sh dss --long)

if [[ -z "${DOCKER_URL}" ]]; then
  echo "DOCKER_URL environment variable is not set; building image to interuss-local/dss..."
  docker image build . -t interuss-local/dss

  echo "Building image to interuss-local/dummy-oauth..."
  docker image build . --file cmds/dummy-oauth/Dockerfile -t interuss-local/dummy-oauth

  echo "Building image to interuss-local/db-manager..."
  docker image build . --file cmds/db-manager/Dockerfile -t interuss-local/db-manager

  echo "DOCKER_URL environment variable was not set; built images to interuss-local/dss and interuss-local/dummy-oauth"
else
  echo "Building image ${DOCKER_URL}/dss:${VERSION}"
  docker image build . -t "${DOCKER_URL}/dss:${VERSION}"

  echo "Pushing docker image ${DOCKER_URL}/dss:${VERSION}..."
  docker image push "${DOCKER_URL}/dss:${VERSION}"

  echo "Built and pushed docker image ${DOCKER_URL}/dss:${VERSION}"

  echo "Building image ${DOCKER_URL}/db-manager:${VERSION}"
  docker image build . --file cmds/db-manager/Dockerfile -t "${DOCKER_URL}/db-manager:${VERSION}"

  echo "Pushing docker image ${DOCKER_URL}/db-manager:${VERSION}..."
  docker image push "${DOCKER_URL}/db-manager:${VERSION}"

  echo "Built and pushed docker image ${DOCKER_URL}/db-manager:${VERSION}"

  echo "VAR_DOCKER_IMAGE_NAME: ${DOCKER_URL}/dss:${VERSION}"
  echo "VAR_SCHEMA_MANAGER_IMAGE_NAME: ${DOCKER_URL}/db-manager:${VERSION}"
fi
