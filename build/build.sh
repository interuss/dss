#!/usr/bin/env bash

# This script builds InterUSS docker images and may be run from any working
# directory.  If run without a DOCKER_URL environment variable, it will just
# build images named interuss-local/*.  If DOCKER_URL is present, it will both
# build the versioned dss image and push it to the DOCKER_URL remote.
# If DOCKER_URL is set:
#
# 1) DOCKER_UPDATE_LATEST can be optionally set to `true` in order to publish
# the latest tag along the version.
#
# 2) DOCKER_SIGN can be optionally set to `true` in order to sign the published
# image using sigstore. When ran within the Github Actions CI, the identity of
# the CI workflow will be used through the ID token emitted by GitHub. When ran
# outside of the CI, `cosign` will interactively ask for an authentication
# against a supported identity provider (Google, GitHub or Microsoft at this time).
# If DOCKER_SIGN is `true`, CERT_IDENTITY and CERT_ISSUER must be set in order
# to verify the signature of the published image.

set -eo pipefail

OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")/.."
else
	BASEDIR=$(readlink -e "$(dirname "$0")/..")
fi
cd "${BASEDIR}"

#VERSION=$(./scripts/git/version.sh dss)  # TODO: hardcoded for testing purposes, remove me before merging
VERSION=cosign-test
LATEST_TAG="latest"

if [[ -z "${DOCKER_URL}" ]]; then
  echo "DOCKER_URL environment variable is not set; building image to interuss-local/dss..."
  docker image build . -t interuss-local/dss

  echo "Building image to interuss-local/dummy-oauth..."
  docker image build . --file cmds/dummy-oauth/Dockerfile -t interuss-local/dummy-oauth

  echo "DOCKER_URL environment variable was not set; built images to interuss-local/dss and interuss-local/dummy-oauth"
else
  TAG="${DOCKER_URL}/dss:${VERSION}"

  echo "Building image ${TAG}"
  docker image build . -t "${TAG}"

  echo "Pushing docker image ${TAG}..."
  docker image push "${TAG}"

  echo "Built and pushed docker image ${TAG}"

  if [[ "${DOCKER_SIGN}" == "true" ]]; then
    # We sign only the first digest of the image. We don't expect multiple ones as we are building for a single architecture.
    DIGEST=$(docker image inspect --format='{{index .RepoDigests 0}}' "${TAG}")
    echo "Signing docker image ${TAG} (digest: ${DIGEST})..."
    cosign sign --yes "${DIGEST}"

    echo "Verifying signature of docker image ${TAG} (digest: ${DIGEST})..."
    cosign verify "${DIGEST}" --certificate-identity="${CERT_IDENTITY}" --certificate-oidc-issuer="${CERT_ISSUER}"

    echo "Signed and verified signature of docker image ${TAG} (digest: ${DIGEST})..."

  fi

  if [[ "${DOCKER_UPDATE_LATEST}" == "true" ]]; then

    echo "Tagging docker image ${DOCKER_URL}/dss:${LATEST_TAG}..."
    docker tag "${TAG}" "${DOCKER_URL}/dss:${LATEST_TAG}"

    echo "Pushing docker image ${DOCKER_URL}/dss:${LATEST_TAG}..."
    docker image push "${DOCKER_URL}/dss:${LATEST_TAG}"

    echo "Built and pushed docker image ${DOCKER_URL}/dss:${LATEST_TAG}"

  fi

  echo "VAR_DOCKER_IMAGE_NAME: ${TAG}"
fi
