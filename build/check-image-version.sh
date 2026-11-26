#!/usr/bin/env bash

# Fails if the dss image given as argument reports a dirty version.

set -eo pipefail

IMAGE=${1:?Usage: $0 <IMAGE>}

IMAGE_LOGS=$(docker run --rm "${IMAGE}" core-service 2>&1 || true)
if grep -q -- "-dirty" <<< "${IMAGE_LOGS}"; then
  echo "Image ${IMAGE} reports a dirty version:" >&2
  echo "${IMAGE_LOGS}" >&2
  exit 1
fi
