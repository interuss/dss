#!/usr/bin/env bash

set -eo pipefail

# This script will run db-manager targeting the CRDB container created by
# run_locally.sh.  See standalone_instance.md for more information.

if [[ -z $(command -v docker) ]]; then
  echo "docker is required but not installed.  Visit https://docs.docker.com/get-docker/ to install."
  exit 1
fi

if [[ -z ${1} ]]; then
  echo "Usage: ${0} <rid|scd> [DB version]"
  echo "  Example: ${0} rid 3.1.1"
  echo "  Example: ${0} rid"
  echo "  Example: ${0} scd latest"
  exit 1
fi

OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi

cd "${BASEDIR}/../.." || exit 1

if [[ -z ${2} ]]; then
  DBVERSION_FLAG=""
else
  DBVERSION_FLAG="--db_version ${2}"
fi

pwd
docker image build . -t interuss-local/dss . || exit 1
# shellcheck disable=SC2086
#                    ^ DBVERSION_FLAG should word-split
docker container run \
    -v "$(pwd)"/build/dev/local-dss-data:/var/local-dss-data \
    --network dss_sandbox-default \
    interuss-local/dss \
         --schemas_dir /db-schemas/"${1}" \
         ${DBVERSION_FLAG} \
         --datastore_host local-dss-crdb || exit 1
