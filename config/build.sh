#!/bin/sh

if [[ -z "${DOCKER_URL}" ]]; then
  echo "Set the DOCKER_URL environment variable before running this script."
  exit 1
fi

BASEDIR=$(readlink -e "$(dirname "$0")/..")

set -e -x

VERSION="$(date -u +%Y-%m-%d)-$(git rev-parse --short HEAD)"

cd "${BASEDIR}"
docker build -f cmds/http-gateway/Dockerfile  . -t $DOCKER_URL/http-gateway:$VERSION
docker build -f cmds/grpc-backend/Dockerfile  . -t $DOCKER_URL/grpc-backend:$VERSION

docker push $DOCKER_URL/http-gateway:$VERSION
docker push $DOCKER_URL/grpc-backend:$VERSION
