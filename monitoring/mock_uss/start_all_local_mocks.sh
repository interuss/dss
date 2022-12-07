#!/bin/bash

set -eo pipefail

# Find and change to repo root directory
OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi
cd "${BASEDIR}/../.." || exit 1

monitoring/mock_uss/run_locally_scdsc.sh -d
export DO_NOT_BUILD_MONITORING=true
monitoring/mock_uss/run_locally_ridsp.sh -d
monitoring/mock_uss/run_locally_riddp.sh -d
monitoring/mock_uss/run_locally_geoawareness.sh -d
monitoring/mock_uss/wait_for_mock_uss.sh mock_uss_scdsc
monitoring/mock_uss/wait_for_mock_uss.sh mock_uss_ridsp
monitoring/mock_uss/wait_for_mock_uss.sh mock_uss_riddp
monitoring/mock_uss/wait_for_mock_uss.sh mock_uss_geoawareness
