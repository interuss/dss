#!/usr/bin/env bash

set -eo pipefail

# This script will deploy a standalone DSS instance with docker-compose.  See
# standalone_instance.md for more information.

if [[ -z $(command -v docker-compose) ]]; then
  echo "docker-compose is required but not installed.  Visit https://docs.docker.com/compose/install/ to install."
  exit 1
fi

OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi

cd "${BASEDIR}" || exit 1

DC_COMMAND=$*

if [[ ! "$DC_COMMAND" ]]; then
  DC_COMMAND="up"
  DC_OPTIONS="--build"
elif [[ "$DC_COMMAND" == "down" ]]; then
  DC_OPTIONS="--volumes --remove-orphans"
elif [[ "$DC_COMMAND" == "debug" ]]; then
  DC_COMMAND=up
  export DEBUG_ON=1
fi

# shellcheck disable=SC2086
docker-compose -f docker-compose_dss.yaml -p dss_sandbox $DC_COMMAND $DC_OPTIONS
