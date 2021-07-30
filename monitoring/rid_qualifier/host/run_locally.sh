#!/usr/bin/env bash

# Find and change to repo root directory
OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi
cd "${BASEDIR}/../../.." || exit 1

# Run monitoring/rid_qualifier/mock/run_locally.sh to produce a mock RID system
# for use with rid_qualifier
AUTH="DummyOAuth(http://host.docker.internal:8085/token,uss1)"
DSS="http://host.docker.internal:8082"
AUD="host.docker.internal"
PORT=8072

docker build \
  -t local-interuss/rid-host \
  -f monitoring/rid_qualifier/host/Dockerfile \
  monitoring \
  || exit 1

# TODO: Cleaning Redis may not be required.
echo "cleaning up any Redis containers"
docker container stop redis &> /dev/null || echo "No Redis server running"
echo "Start Redis container"
docker run --name redis --rm -d -p 6379:6379 redis:3-alpine

echo "cleaning up any RQ worker containers"
docker container stop rq-worker &> /dev/null || echo "No RQ worker running"
echo "Start RQ worker container."
docker run --name rq-worker -d --rm \
  -e MOCK_HOST_AUTH_SPEC="${AUTH}" \
  -e MOCK_HOST_DSS_URL="${DSS}" \
  -e REDIS_URL=redis://redis-server:6379/0 \
  -v `pwd`/build/test-certs:/var/test-certs:ro \
  -v /tmp/rid-host-input-files:/mnt/app/input-files \
  --link redis:redis-server \
  --entrypoint /usr/local/bin/rq \
  local-interuss/rid-host \
  worker -u redis://redis-server:6379/0 qualifer-tasks

echo "Start Host container"
docker run --name rid-host \
  --rm \
  -e MOCK_HOST_AUTH_SPEC="${AUTH}" \
  -e MOCK_HOST_DSS_URL="${DSS}" \
  -e REDIS_URL=redis://redis-server:6379/0 \
  -p ${PORT}:5000 \
  -v `pwd`/build/test-certs:/var/test-certs:ro \
  -v /tmp/rid-host-input-files:/mnt/app/input-files \
  --link redis:redis-server \
  local-interuss/rid-host \
  gunicorn \
    --preload \
    --workers=1 \
    --bind=0.0.0.0:5000 \
    monitoring.rid_qualifier.host:webapp
