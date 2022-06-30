#!/usr/bin/env bash

set -eo pipefail

echo "Re/Create probe_local_instance_test_result file"
RESULTFILE="$(pwd)/probe_local_instance_test_result.xml"
touch "${RESULTFILE}"
GATEWAY_CONTAINER="dss_sandbox_local-dss-http-gateway_1"
OAUTH_CONTAINER="dss_sandbox_local-dss-dummy-oauth_1"
declare -a localhost_containers=("$GATEWAY_CONTAINER" "$OAUTH_CONTAINER")

for container_name in "${localhost_containers[@]}"; do
	if [ "$( docker container inspect -f '{{.State.Status}}' "$container_name" )" == "running" ]; then
		echo "$container_name available!"
	else
		echo "Error: $container_name not running. Execute 'run_locally.sh up' before running probe_local_instance.sh";
		exit 1;
	fi
done

OS=$(uname)
if [[ $OS == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")/../.."
else
	BASEDIR=$(readlink -e "$(dirname "$0")/../..")
fi

echo "probe_local_instance base directory is ${BASEDIR}"
cd "${BASEDIR}"

sleep 1
echo " -------------- PYTEST -------------- "
echo "Building Integration Test container"
pwd
docker build -q --rm -f monitoring/Dockerfile monitoring -t probe-local-test

echo "Finally Begin Testing"
docker run --network dss_sandbox_default \
  --link $OAUTH_CONTAINER:oauth \
	--link $GATEWAY_CONTAINER:local-gateway \
	-v "${RESULTFILE}:/app/test_result" \
	-v "$(pwd):/app" \
	-w /app/monitoring/prober \
	probe-local-test \
	pytest /app/monitoring/prober	\
	--junitxml=/app/test_result \
	--dss-endpoint http://local-gateway:8082 \
	--rid-auth "DummyOAuth(http://oauth:8085/token,sub=fake_uss)" \
	--scd-auth1 "DummyOAuth(http://oauth:8085/token,sub=fake_uss)" \
	--scd-auth2 "DummyOAuth(http://oauth:8085/token,sub=fake_uss2)"	\
	--scd-api-version 1.0.0
