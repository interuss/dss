#!/bin/bash

# Find and change to repo root directory
OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi
cd "${BASEDIR}/../.." || exit 1

docker container rm -f mock_uss_scdsc mock_uss_ridsp mock_uss_riddp mock_uss_geoawareness mock_uss_atproxy_client
