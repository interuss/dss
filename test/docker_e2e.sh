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
	docker logs core-service-for-testing 2> core-service-for-testing.log
}

function cleanup() {
	# ----------- clean up -----------
	echo "Stopping dummy oauth container"
	docker rm -f dummy-oauth-for-testing &> /dev/null || true

	echo "Stopping core-service container"
	docker kill -f core-service-for-testing &> /dev/null || true

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
echo "Building local container for testing (see core-service-build.log for details)"
docker build --rm . -t local-interuss-dss-image > core-service-build.log

echo " ---------------- CRDB -------------------- "
echo "cleaning up any crdb pre-existing containers"
docker rm -f dss-crdb-for-debugging &> /dev/null || echo "No CRDB to clean up"

echo "Starting cockroachdb with admin port on :8080"
docker run -d --rm --name dss-crdb-for-debugging \
	-p 26257:26257 \
	-p 8080:8080 \
	cockroachdb/cockroach:v21.2.7 start-single-node \
	--insecure > /dev/null

sleep 1
echo "Bootstrapping RID Database tables"
docker run --rm --name rid-db-manager \
	--link dss-crdb-for-debugging:crdb \
	-v "$(pwd)/build/deploy/db_schemas/rid:/db-schemas/rid" \
	local-interuss-dss-image \
	/usr/bin/db-manager \
	--schemas_dir db-schemas/rid \
	--db_version "latest" \
	--cockroach_host crdb

sleep 1
echo "Bootstrapping SCD Database tables"
docker run --rm --name scd-db-manager \
	--link dss-crdb-for-debugging:crdb \
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

echo "Starting core service on :8082"
docker run -d --name core-service-for-testing \
	--link dss-crdb-for-debugging:crdb \
	-v "$(pwd)/build/test-certs/auth2.pem:/app/test.crt" \
	local-interuss-dss-image \
	core-service \
    -addr :8082 \
	--cockroach_host crdb \
	-public_key_files /app/test.crt \
	-log_format console \
	-dump_requests \
	-accepted_jwt_audiences core-service \
	-enable_scd	\
	-enable_http

echo " -------------- DUMMY OAUTH -------------- "
echo "Building dummy-oauth server container"
docker build --rm -f cmds/dummy-oauth/Dockerfile . -t local-dummy-oauth > dummy-oauth-build.log

echo "Cleaning up any pre-existing dummy-oauth container"
docker rm -f dummy-oauth-for-testing &> /dev/null || echo "No dummy oauth to clean up"

echo "Starting mock oauth server on :8085"
docker run -d --name dummy-oauth-for-testing -p 8085:8085 \
	-v "$(pwd)/build/test-certs/auth2.key:/app/test.key" \
	local-dummy-oauth \
	-private_key_file /app/test.key

sleep 1
echo " -------------- PYTEST -------------- "
echo "Building monitoring (Integration Test) image"
docker build -q --rm -f monitoring/Dockerfile monitoring -t interuss/monitoring

echo "Finally Begin Testing"
if ! docker run --link dummy-oauth-for-testing:oauth \
	--link core-service-for-testing:core-service \
	-v "${RESULTFILE}:/app/test_result" \
	-w /app/monitoring/prober \
	interuss/monitoring \
	pytest \
	"${1:-.}" \
	-rsx \
	--junitxml=/app/test_result \
	--dss-endpoint http://core-service:8082 \
	--rid-auth "DummyOAuth(http://oauth:8085/token,sub=fake_uss)" \
	--rid-v2-auth "DummyOAuth(http://oauth:8085/token,sub=fake_uss)" \
	--scd-auth1 "DummyOAuth(http://oauth:8085/token,sub=fake_uss)" \
	--scd-auth2 "DummyOAuth(http://oauth:8085/token,sub=fake_uss2)"	\
	--scd-api-version 1.0.0; then

    if [ "$CI" == "true" ]; then
        echo "=== END OF TEST RESULTS ==="
        echo "Dumping core-service logs"
        docker logs core-service-for-testing
    fi
fi

echo "Cleaning up core-service container"
docker stop core-service-for-testing > /dev/null
test "$(docker inspect core-service-for-testing --format='{{.State.ExitCode}}')" = 0
