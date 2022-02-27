#!/usr/bin/env bash

set -eo pipefail

# This script is to perform simple read operations on SCD database for testing.

OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi

cd "${BASEDIR}" || exit 1

echo "DSS response to [SCD] GET subscription query:"
echo "============="
./read_scd_subscription.sh 00000158-9aba-4026-bbef-bad6cde80000
echo

echo "DSS response to [SCD] GET constraint reference query:"
echo "============="
./read_scd_constraint_reference.sh 00000159-9aba-4026-bbef-bad6cde80000
echo

echo "DSS response to [SCD] GET operational intent reference query:"
echo "============="
./read_scd_operational_intent_reference.sh 0000015a-9aba-4026-bbef-bad6cde80000
echo
