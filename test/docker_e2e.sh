#!/bin/bash

set -eo pipefail

echo "Re/Create e2e_test_result file"
RESULTFILE="$(pwd)/e2e_test_result"
touch $RESULTFILE
cat /dev/null > $RESULTFILE

OS=$(uname)
if [[ $OS == "Darwin" ]]; then 
	# OSX uses BSD readlink
	BASEDIR="$(dirname $0)/.."
else
	BASEDIR=$(readlink -e "$(dirname "$0")/..")
fi

cd "${BASEDIR}"

echo "cleaning up any crdb pre-existing containers"
docker stop dss-crdb-for-debugging || echo "No CRDB to clean up"

echo "Starting cockroachdb with admin port on :8080"
docker run -d --rm --name dss-crdb-for-debugging \
	-p 26257:26257 \
	-p 8080:8080 \
	cockroachdb/cockroach:v19.1.2 start \
	--insecure > /dev/null

sleep 5
echo " -------------- BOOTSTRAP ----------------- "
echo "Building local container for testing"
docker build -q --rm . -t local-interuss-dss-image

echo " ------------ GRPC BACKEND ---------------- "
echo "Cleaning up any pre-existing grpc-backend container"
docker stop grpc-backend-for-testing || echo "No grpc backend to clean up"

echo "Starting grpc backend on :8081"
docker run -d --rm --name grpc-backend-for-testing \
	--link dss-crdb-for-debugging:crdb \
	-v $(pwd)/config/test-certs/auth2.pem:/app/test.crt \
	local-interuss-dss-image \
	grpc-backend \
	--cockroach_host crdb \
	-public_key_file /app/test.crt \
	-reflect_api \
	-log_format console \
	-dump_requests \
	-jwt_audience local-gateway

docker logs -f grpc-backend-for-testing 2> grpc-backend-for-testing.log &

sleep 5
echo " ------------- HTTP GATEWAY -------------- "
echo "Cleaning up any pre-existing http-gateway container"
docker stop http-gateway-for-testing || echo "No http gateway to clean up"

echo "Starting http-gateway on :8082"
docker run -d --rm --name http-gateway-for-testing -p 8082:8082 \
	--link grpc-backend-for-testing:grpc \
	local-interuss-dss-image \
	http-gateway \
	-grpc-backend grpc:8081 \
	-addr :8082 \
	-trace-requests

docker logs -f http-gateway-for-testing 2> http-gateway-for-testing.log &

sleep 5
echo " -------------- DUMMY OAUTH -------------- "
echo "Building dummy-oauth server container"
docker build -q --rm -f cmds/dummy-oauth/Dockerfile . -t local-dummy-oauth

echo "Cleaning up any pre-existing dummy-oauth container"
docker stop dummy-oauth-for-testing || echo "No dummy oauth to clean up"

echo "Starting mock oauth server on : 8085"
docker run -d --rm --name dummy-oauth-for-testing -p 8085:8085 \
	-v $(pwd)/config/test-certs/auth2.key:/app/test.key \
	local-dummy-oauth \
	-private_key_file /app/test.key

sleep 5
echo " -------------- PYTEST -------------- "
echo "Building Integration Test container"
docker build -q --rm -f monitoring/prober/Dockerfile monitoring/prober -t e2e-test

echo "Finally Begin Testing"
docker run --link dummy-oauth-for-testing:oauth \
	--link http-gateway-for-testing:local-gateway \
	-v $RESULTFILE:/app/test_result \
	e2e-test \
	--junitxml=/app/test_result \
	--oauth-token-endpoint http://oauth:8085/token \
	--dss-endpoint http://local-gateway:8082 \
	--use-dummy-oauth 1 \
	--api-version-role '/v1/dss' \
	-vv

# ----------- clean up -----------
echo "Stopping dummy oauth container"
docker stop dummy-oauth-for-testing

echo "Stopping http gateway container"
docker stop http-gateway-for-testing

echo "Stopping grpc-backend container"
docker stop grpc-backend-for-testing

echo "Stopping crdb docker"
docker stop dss-crdb-for-debugging

jobs -p | xargs kill