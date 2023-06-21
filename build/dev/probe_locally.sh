#!/usr/bin/env bash

set -eo pipefail
set -x

# Find and change to repo root directory
OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi
cd "${BASEDIR}/../.." || exit 1

CORE_SERVICE_CONTAINER="dss_sandbox-local-dss-core-service-1"
OAUTH_CONTAINER="dss_sandbox-local-dss-dummy-oauth-1"
declare -a localhost_containers=("$CORE_SERVICE_CONTAINER" "$OAUTH_CONTAINER")

for container_name in "${localhost_containers[@]}"; do
	if [ "$( docker container inspect -f '{{.State.Status}}' "$container_name" )" == "running" ]; then
		echo "$container_name available!"
	else
    echo '#########################################################################'
    echo '## Prerequisite to run this command is:                                ##'
    echo '## Local DSS instance + Dummy OAuth server (/build/dev/run_locally.sh) ##'
    echo '#########################################################################'
		echo "Error: $container_name not running. Execute 'build/dev/run_locally.sh up' before running build/dev/probe_locally.sh";
		exit 1;
	fi
done

echo "Re/Create e2e_test_result file"
RESULTFILE="$(pwd)/e2e_test_result"
touch "${RESULTFILE}"
cat /dev/null > "${RESULTFILE}"

if ! docker run --link "$OAUTH_CONTAINER":oauth \
	--link "$CORE_SERVICE_CONTAINER":core-service \
	--network dss_sandbox_default \
	-v "${RESULTFILE}:/app/test_result" \
	-w /app/monitoring/prober \
	interuss/monitoring:v0.2.0 \
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
        docker logs "$CORE_SERVICE_CONTAINER"
    fi
    echo "Prober did not succeed."
    exit 1
else
    echo "Prober succeeded."
fi
