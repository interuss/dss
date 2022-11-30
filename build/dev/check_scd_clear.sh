#!/usr/bin/env bash

set -eo pipefail

# This script is to clean up after check operations on SCD database for testing.

OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi

cd "${BASEDIR}" || exit 1

# Retrieve token from dummy OAuth server
ACCESS_TOKEN=$(curl --silent \
    "http://localhost:8085/token?grant_type=client_credentials&scope=utm.strategic_coordination%20utm.constraint_management&intended_audience=localhost&issuer=localhost&sub=check_scd" \
| python extract_json_field.py access_token)


echo "DSS response to [SCD] DELETE subscription query:"
echo "============="

VERSION=$(./read_scd_subscription.sh 00000158-9aba-4026-bbef-bad6cde80000 | python extract_json_field.py 'subscription.version')

curl --silent -X DELETE  \
"http://localhost:8082/dss/v1/subscriptions/00000158-9aba-4026-bbef-bad6cde80000/${VERSION}" \
-H "Authorization: Bearer ${ACCESS_TOKEN}" -H "Content-Type: application/json"
echo


echo "DSS response to [SCD] DELETE constraint reference query:"
echo "============="

VERSION=$(./read_scd_constraint_reference.sh 00000159-9aba-4026-bbef-bad6cde80000 | python extract_json_field.py 'constraint_reference.version')

curl --silent -X DELETE  "http://localhost:8082/dss/v1/constraint_references/00000159-9aba-4026-bbef-bad6cde80000/${VERSION}"  \
-H "Authorization: Bearer ${ACCESS_TOKEN}"  \
-H "Content-Type: application/json"
echo


echo "DSS response to [SCD] DELETE operational intent reference query:"
echo "=========="

VERSION=$(./read_scd_operational_intent_reference.sh 0000015a-9aba-4026-bbef-bad6cde80000 | python extract_json_field.py 'operational_intent_reference.version')

curl --silent -X DELETE  "http://localhost:8082/dss/v1/operational_intent_references/0000015a-9aba-4026-bbef-bad6cde80000/${VERSION}"  \
 -H "Authorization: Bearer ${ACCESS_TOKEN}"  \
 -H "Content-Type: application/json"
echo
