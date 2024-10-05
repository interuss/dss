#!/usr/bin/env bash

set -eo pipefail

echo "Run Post migration Setup"
CRDB_MIGRATION_CONTAINER="dss-crdb-for-migration-testing"

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
	docker rm -f $CRDB_MIGRATION_CONTAINER &> /dev/null || true
}

function on_exit() {
	gather_logs || true
	cleanup
}

trap on_exit   EXIT

if [ "$( docker container inspect -f '{{.State.Status}}' "$CRDB_MIGRATION_CONTAINER" )" == "running" ]; then
    echo "$CRDB_MIGRATION_CONTAINER available!"
else
    echo "Error: $CRDB_MIGRATION_CONTAINER not running. Execute 'clear_db.sh and migrate_db.sh' before running post_migration_e2e.sh";
    exit 1;
fi

echo "Bootstrapping SCD Database tables"
docker run --rm --name scd-db-manager \
	--link $CRDB_MIGRATION_CONTAINER:crdb \
	local-db-manager \
	--schemas_dir db-schemas/scd \
	--db_version "latest" \
	--cockroach_host crdb

sleep 1
echo " ------------ CORE SERVICE ---------------- "
echo "Cleaning up any pre-existing core-service container"
docker rm -f core-service-for-testing &> /dev/null || echo "No core service to clean up"

echo "Starting core service on :8082"
docker run -d --name core-service-for-testing \
	--link $CRDB_MIGRATION_CONTAINER:crdb \
	-v "$(pwd)/build/test-certs/auth2.pem:/app/test.crt" \
	local-interuss-dss-image \
	core-service \
	--cockroach_host crdb \
	-public_key_files /app/test.crt \
	-log_format console \
	-dump_requests \
	-accepted_jwt_audiences core-service \
	-enable_scd	\
	-allow_http_base_urls

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
echo "Finally Begin Testing"
docker run --link dummy-oauth-for-testing:oauth \
	--link core-service-for-testing:core-service \
	-v "${RESULTFILE}:/app/test_result" \
	-w /app/monitoring/prober \
	interuss/monitoring:v0.11.1 \
	pytest \
	"${1:-.}" \
	-rsx \
	--junitxml=/app/test_result \
	--dss-endpoint http://core-service:8082 \
	--rid-auth "DummyOAuth(http://oauth:8085/token,sub=fake_uss)" \
	--scd-auth1 "DummyOAuth(http://oauth:8085/token,sub=fake_uss)" \
	--scd-auth2 "DummyOAuth(http://oauth:8085/token,sub=fake_uss2)"	\
	--scd-api-version 1.0.0

echo "Cleaning up core-service container"
docker stop core-service-for-testing > /dev/null
test "$(docker inspect core-service-for-testing --format='{{.State.ExitCode}}')" = 0
