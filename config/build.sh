#!/bin/sh

if [[ -z "${CLOUD_PROJECT}" ]]; then
  echo "Set the CLOUD_PROJECT environment variable before running this script."
  exit 1
fi

BASEDIR=$(readlink -e "$(dirname "$0")/..")

set -e -x

VERSION="$(date -u +%Y-%m-%d)-$(git rev-parse --short HEAD)"

cd "${BASEDIR}"
docker build -f cmds/http-gateway/Dockerfile  . -t gcr.io/$CLOUD_PROJECT/http-gateway:$VERSION
docker build -f cmds/grpc-backend/Dockerfile  . -t gcr.io/$CLOUD_PROJECT/grpc-backend:$VERSION

docker push gcr.io/$CLOUD_PROJECT/http-gateway:$VERSION
docker push gcr.io/$CLOUD_PROJECT/grpc-backend:$VERSION
