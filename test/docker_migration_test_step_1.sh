#!/usr/bin/env bash

set -eo pipefail

echo "Re/Create e2e_test_result file"
RESULTFILE="$(pwd)/e2e_test_result"
touch "${RESULTFILE}"
cat /dev/null > "${RESULTFILE}"

OS=$(uname)
if [[ $OS == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")/.."
else
	BASEDIR=$(readlink -e "$(dirname "$0")/..")
fi

echo "e2e base directory is ${BASEDIR}"
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
	# cleanup
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
echo "Building db-manager container for testing"
docker build --rm -f cmds/db-manager/Dockerfile . -t local-db-manager > db-manager-build.log

echo " ---------------- CRDB -------------------- "
echo "cleaning up any crdb pre-existing containers"
docker rm -f dss-crdb-for-debugging &> /dev/null || echo "No CRDB to clean up"

echo "Starting cockroachdb with admin port on :8080"
docker run -d --rm --name dss-crdb-for-debugging \
	-p 26257:26257 \
	-p 8080:8080 \
	cockroachdb/cockroach:v20.2.0 start-single-node \
	--insecure > /dev/null

sleep 1
echo "Bootstrapping RID Database tables"
docker run --rm --name rid-db-manager \
	--link dss-crdb-for-debugging:crdb \
	-v "$(pwd)/build/deploy/db_schemas/rid:/db-schemas/rid" \
	local-db-manager \
	--schemas_dir db-schemas/rid \
	--db_version "3.1.1" \
	--migration_step 7	\
	--cockroach_host crdb
