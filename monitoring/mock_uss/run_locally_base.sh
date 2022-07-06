#!/usr/bin/env bash

# Find and change to repo root directory
OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi
cd "${BASEDIR}/../.." || exit 1

echo '#########################################################################'
echo '## NOTE: Prerequisite to run this command is:                          ##'
echo '## Local DSS instance + Dummy OAuth server (/build/dev/run_locally.sh) ##'
echo '#########################################################################'

monitoring/build.sh || exit 1
