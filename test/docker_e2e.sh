#!/usr/bin/env bash

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

function gather_logs() {
	docker logs http-gateway-for-testing 2> http-gateway-for-testing.log
	docker logs grpc-backend-for-testing 2> grpc-backend-for-testing.log
}

function cleanup() {
	# ----------- clean up -----------
	echo "Stopping dummy oauth container"
	docker rm -f dummy-oauth-for-testing &> /dev/null || true

	echo "Stopping http gateway container"
	docker kill -f http-gateway-for-testing &> /dev/null || true

	echo "Stopping grpc-backend container"
	docker kill -f grpc-backend-for-testing &> /dev/null || true

	echo "Stopping crdb docker"
	docker rm -f dss-crdb-for-debugging &> /dev/null || true
}

function on_exit() {
	gather_logs || true
	cleanup
}

function on_sigint() {
	cleanup
	exit
}

trap on_exit   EXIT
trap on_sigint SIGINT


echo " -------------- BOOTSTRAP ----------------- "
echo "Building local container for testing (see grpc-backend-build.log for details)"
docker build --rm . -t local-interuss-dss-image > grpc-backend-build.log

echo " ---------------- CRDB -------------------- "
echo "cleaning up any crdb pre-existing containers"
docker rm -f dss-crdb-for-debugging &> /dev/null || echo "No CRDB to clean up"

echo "Starting cockroachdb with admin port on :8080"
docker run -d --rm --name dss-crdb-for-debugging \
	-p 26257:26257 \
	-p 8080:8080 \
	cockroachdb/cockroach:v20.1.1 start \
	--insecure > /dev/null

sleep 1
echo " ------------ GRPC BACKEND ---------------- "
echo "Cleaning up any pre-existing grpc-backend container"
docker rm -f grpc-backend-for-testing &> /dev/null || echo "No grpc backend to clean up"

echo "Starting grpc backend on :8081"
docker run -d --name grpc-backend-for-testing \
	--link dss-crdb-for-debugging:crdb \
	-v $(pwd)/build/test-certs/auth2.pem:/app/test.crt \
	-e DSS_ERRORS_OBFUSCATE_INTERNAL_ERRORS='false' \
	local-interuss-dss-image \
	grpc-backend \
	--cockroach_host crdb \
	-public_key_files /app/test.crt \
	-reflect_api \
	-log_format console \
	-dump_requests \
	-accepted_jwt_audiences local-gateway \
	-enable_scd

sleep 1
echo " ------------- HTTP GATEWAY -------------- "
echo "Cleaning up any pre-existing http-gateway container"
docker rm -f http-gateway-for-testing &> /dev/null || echo "No http gateway to clean up"

echo "Starting http-gateway on :8082"
docker run -d --name http-gateway-for-testing -p 8082:8082 \
	--link grpc-backend-for-testing:grpc \
	local-interuss-dss-image \
	http-gateway \
	-grpc-backend grpc:8081 \
	-addr :8082 \
	-trace-requests \
	-enable_scd

echo " -------------- DUMMY OAUTH -------------- "
echo "Building dummy-oauth server container"
docker build -q --rm -f cmds/dummy-oauth/Dockerfile . -t local-dummy-oauth

echo "Cleaning up any pre-existing dummy-oauth container"
docker rm -f dummy-oauth-for-testing &> /dev/null || echo "No dummy oauth to clean up"

echo "Starting mock oauth server on :8085"
docker run -d --name dummy-oauth-for-testing -p 8085:8085 \
	-v $(pwd)/build/test-certs/auth2.key:/app/test.key \
	local-dummy-oauth \
	-private_key_file /app/test.key

sleep 1
echo " -------------- PYTEST -------------- "
echo "Building Integration Test container"
docker build -q --rm -f monitoring/prober/Dockerfile monitoring/prober -t e2e-test

echo "Finally Begin Testing"
docker run --link dummy-oauth-for-testing:oauth \
	--link http-gateway-for-testing:local-gateway \
	-v $RESULTFILE:/app/test_result \
	e2e-test \
	${1:-.} \
	--junitxml=/app/test_result \
	--oauth-token-endpoint http://oauth:8085/token \
	--dss-endpoint http://local-gateway:8082 \
	--use-dummy-oauth 1 \
	--api-version-role '/v1/dss' \
	--scd-dss-endpoint http://local-gateway:8082/dss/v1 \
	-vv
