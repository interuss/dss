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

echo "========== Running legacy DSS prober =========="
if ! docker run --link "$OAUTH_CONTAINER":oauth \
	--link "$CORE_SERVICE_CONTAINER":core-service \
	--network dss_sandbox-default \
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
        echo "=== END OF LEGACY PROBER TEST RESULTS ==="
        echo "Dumping core-service logs"
        docker logs "$CORE_SERVICE_CONTAINER"
    fi
    echo "Prober did not succeed."
    exit 1
else
    echo "Prober succeeded."
fi

# TODO: ugly non-optimized way of getting the code we need
git clone --recurse-submodules --branch uq_dss_probing git@github.com:Orbitalize/monitoring.git monitoring-repo-tmp
pushd monitoring-repo-tmp
make build-monitoring
popd
rm -rf monitoring-repo-tmp

USS_QUALIFIER_CONF="$(pwd)/build/dev/probe_locally_configuration.yaml"
OUTPUT_DIR="$(pwd)/build/dev/probe_locally_output"
mkdir -p "$OUTPUT_DIR"

echo "========== Running uss_qualifier for DSS probing =========="
# shellcheck disable=SC2086
docker run --name dss_probing \
  --rm \
  --link "$OAUTH_CONTAINER":oauth \
  --link "$CORE_SERVICE_CONTAINER":core-service \
  --network dss_sandbox-default \
  -u "$(id -u):$(id -g)" \
  -e PYTHONBUFFERED=1 \
  -e AUTH_SPEC='DummyOAuth(http://oauth:8085/token,sub=fake_uss)' \
  -e USS_QUALIFIER_STOP_FAST=true \
  -v "${OUTPUT_DIR}:/app/monitoring/uss_qualifier/output" \
  -v "${USS_QUALIFIER_CONF}:/app/monitoring/uss_qualifier/configurations/dev/dss_probing.yaml" \
  -w /app/monitoring/uss_qualifier \
  interuss/monitoring \
  python main.py --config configurations.dev.dss_probing

# Set return code according to whether the test run was fully successful
successful=$(jq '.report | .[] | .successful' "${OUTPUT_DIR}/report_dss_probing.json")
if echo "${successful}" | grep -iqF true; then
  echo "Full success indicated by DSS probing"
else
  echo "Could not establish that the DSS probing passed"
  if [ "$CI" == "true" ]; then
    echo "=== END OF USS QUALIFIER TEST RESULTS ==="
    echo "Dumping core-service logs"
    docker logs "$CORE_SERVICE_CONTAINER"
  fi
  exit 1
fi
