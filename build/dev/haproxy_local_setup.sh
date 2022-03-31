#!/usr/bin/env bash
# shellcheck disable=SC2086

set -eo pipefail

echo "DSS/crdb cluster setup  using HAProxy"
RESULTFILE="$(pwd)/haproxy_cluster_setup"
touch "${RESULTFILE}"
cat /dev/null > "${RESULTFILE}"
FLAGS="--network dss_sandbox_default cockroachdb/cockroach:v21.2.3 start --insecure --join=roacha,roachb,roachc"

OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi

cd "${BASEDIR}/../.." || exit 1

DC_COMMAND=$*

function cleanup() {
	# ----------- clean up -----------
	echo "Stopping node roacha container"
	docker rm -f roacha &> /dev/null || true

	echo "Stopping node roachb container"
	docker rm -f roachb &> /dev/null || true

	echo "Stopping node roachc container"
	docker rm -f roachc &> /dev/null || true
	
	echo "Stopping haproxy docker"
	docker rm -f dss-crdb-cluster-for-testing &> /dev/null || true

	echo "Stopping dummy oauth container"
	docker rm -f dummy-oauth-for-testing &> /dev/null || true

	echo "Stopping http gateway container"
	docker kill -f http-gateway-for-testing &> /dev/null || true
	docker rm -f http-gateway-for-testing &> /dev/null || true

	echo "Stopping core-service container"
	docker kill -f core-service-for-testing &> /dev/null || true
	docker rm -f core-service-for-testing &> /dev/null || true

	echo "Removing DSS network"
	docker network rm dss_sandbox_default
}

if [[ "$DC_COMMAND" == "down" ]]; then
  cleanup || true
  exit
fi

echo "HAProxy base directory is ${BASEDIR}"
# cd "${BASEDIR}"

function gather_logs() {
	docker logs http-gateway-for-testing 2> http-gateway-for-testing.log
	docker logs core-service-for-testing 2> core-service-for-testing.log
}


function on_exit() {
	gather_logs || true
}

function on_sigint() {
	cleanup
	exit
}

trap on_exit   EXIT
trap on_sigint SIGINT


echo " -------------- BOOTSTRAP ----------------- "
echo "Building local container for testing (see core-service-build.log for details)"
docker build --rm . -t local-interuss-dss-image > core-service-build.log

echo " ---------------- CRDB -------------------- "
echo "cleaning up any crdb pre-existing containers"
docker rm -f roacha &> /dev/null || echo "No CRDB to clean up"
docker rm -f roachb &> /dev/null || echo "No CRDB to clean up"
docker rm -f roachc &> /dev/null || echo "No CRDB to clean up"

echo "Stopping haproxy container"
docker rm -f dss-crdb-cluster-for-testing &> /dev/null || true

echo "Create DSS network"
docker network create dss_sandbox_default

echo "Starting roacha with admin port on :8080"
docker run -d --rm --name roacha \
	-p 8080:8080 \
	$FLAGS > /dev/null

echo "Starting roachb with admin port on :8088"
docker run -d --rm --name roachb \
	-p 8088:8088 \
	$FLAGS > /dev/null

echo "Starting roachc with admin port on :8089"
docker run -d --rm --name roachc \
	-p 8089:8089 \
	$FLAGS > /dev/null

echo "Initialize cluster setup"
docker exec -it roacha	\
	./cockroach init --insecure

echo "Waiting for cluster to initialize... "
sleep 5

echo "Generate haproxy.cfg from one of the cluster nodes"
docker exec -it roacha ./cockroach gen haproxy --insecure


echo "Copy haproxy.cfg file to a local folder"
docker exec -it roacha cat haproxy.cfg > "$(pwd)/haproxy.cfg"

echo "Start the HAProxy container by mounting the cfg file."
docker run -d --name dss-crdb-cluster-for-testing	\
	--network dss_sandbox_default	\
	-p 26257:26257	\
	-v "$(pwd)/haproxy.cfg:/usr/local/etc/haproxy/haproxy.cfg":ro haproxy:1.7  
sleep 1

echo "Bootstrapping RID Database tables"
docker run --rm --name rid-db-manager \
	--link dss-crdb-cluster-for-testing:crdb \
	--network dss_sandbox_default	\
	-v "$(pwd)/build/deploy/db_schemas/rid:/db-schemas/rid" \
	local-interuss-dss-image \
	/usr/bin/db-manager \
	--schemas_dir db-schemas/rid \
	--db_version "latest" \
	--cockroach_host crdb

sleep 1
echo "Bootstrapping SCD Database tables"
docker run --rm --name scd-db-manager \
	--link dss-crdb-cluster-for-testing:crdb \
	--network dss_sandbox_default	\
	-v "$(pwd)/build/deploy/db_schemas/scd:/db-schemas/scd" \
	local-interuss-dss-image \
	/usr/bin/db-manager \
	--schemas_dir db-schemas/scd \
	--db_version "latest" \
	--cockroach_host crdb

sleep 1
echo " ------------ CORE SERVICE ---------------- "
echo "Cleaning up any pre-existing core-service container"
docker rm -f core-service-for-testing &> /dev/null || echo "No core service to clean up"

echo "Starting core service on :8081"
docker run -d --name core-service-for-testing \
	--link dss-crdb-cluster-for-testing:crdb \
	--network dss_sandbox_default	\
	-v "$(pwd)/build/test-certs/auth2.pem:/app/test.crt" \
	local-interuss-dss-image \
	core-service \
	--cockroach_host crdb \
	-public_key_files /app/test.crt \
	-reflect_api \
	-log_format console \
	-dump_requests \
	-accepted_jwt_audiences local-gateway,localhost \
	-enable_scd	\
	-enable_http

sleep 1
echo " ------------- HTTP GATEWAY -------------- "
echo "Cleaning up any pre-existing http-gateway container"
docker rm -f http-gateway-for-testing &> /dev/null || echo "No http gateway to clean up"

echo "Starting http-gateway on :8082"
docker run -d --name http-gateway-for-testing -p 8082:8082 \
	--link core-service-for-testing:grpc \
	--network dss_sandbox_default	\
	local-interuss-dss-image \
	http-gateway \
	-core-service grpc:8081 \
	-addr :8082 \
	-trace-requests \
	-enable_scd

echo " -------------- DUMMY OAUTH -------------- "
echo "Building dummy-oauth server container"
docker build --rm -f cmds/dummy-oauth/Dockerfile . -t local-dummy-oauth > dummy-oauth-build.log

echo "Cleaning up any pre-existing dummy-oauth container"
docker rm -f dummy-oauth-for-testing &> /dev/null || echo "No dummy oauth to clean up"

echo "Starting mock oauth server on :8085"
docker run -d --name dummy-oauth-for-testing -p 8085:8085 \
	--network dss_sandbox_default	\
	-v "$(pwd)/build/test-certs/auth2.key:/app/test.key" \
	local-dummy-oauth \
	-private_key_file /app/test.key

echo "Finished setting up Local dss cluster."