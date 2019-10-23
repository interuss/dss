#!/bin/sh

if [[ -z "${DOCKER_URL}" ]]; then
  echo "Set the DOCKER_URL environment variable before running this script."
  exit 1
fi

BASEDIR=$(readlink -e "$(dirname "$0")/..")

set -e -x

VERSION="$(date -u +%Y-%m-%d)-$(git rev-parse --short HEAD)"

cd "${BASEDIR}"
docker build . -t $DOCKER_URL/dss:$VERSION

docker push $DOCKER_URL/dss:$VERSION
echo $DOCKER_URL/dss:$VERSION