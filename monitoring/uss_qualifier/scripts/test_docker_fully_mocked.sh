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

echo "Ensure the environment is clean"
echo "============="
make down-locally
make stop-uss-mocks

function cleanup() {
  echo "Clean up"
  echo "============="
  make stop-uss-mocks
  make down-locally
}

function on_exit() {
	cleanup
}

function on_sigint() {
	cleanup
	exit
}

trap on_exit   EXIT
trap on_sigint SIGINT

echo "Start mock system"
echo "============="
make start-locally
make start-uss-mocks

echo "Run the standard local tests."
echo "============="
monitoring/uss_qualifier/run_locally.sh

# Ensure all tests passed
successful=$(python build/dev/extract_json_field.py report.test_suite.successful < monitoring/uss_qualifier/report.json)
if echo "${successful}" | grep -iqF true; then
  echo "All uss_qualifier tests passed."
else
  echo "Could not establish that all uss_qualifier tests passed."
  exit 1
fi
