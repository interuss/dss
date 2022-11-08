#!/usr/bin/env bash

set -eo pipefail
set -o xtrace

# Find and change to repo root directory
OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi
cd "${BASEDIR}/../../.." || exit 1

monitoring/build.sh || exit 1

# shellcheck disable=SC2086
docker run --name test_documentation_formatter \
  --rm \
  -v "$(pwd):/app" \
  interuss/monitoring \
  uss_qualifier/scripts/in_container/format_test_documentation.sh
