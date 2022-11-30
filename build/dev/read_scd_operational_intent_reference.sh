#!/usr/bin/env bash

set -eo pipefail

# Test Get SCD operational intents.

operation_id=$1
[[ -z "$operation_id" ]] && { echo "Error: Operation ID not provided"; exit 1; }

# Retrieve token from dummy OAuth server
ACCESS_TOKEN=$(curl --silent \
    "http://localhost:8085/token?grant_type=client_credentials&scope=utm.strategic_coordination&intended_audience=localhost&issuer=localhost&sub=check_scd" \
| python extract_json_field.py 'access_token')

curl --silent -X GET  "http://localhost:8082/dss/v1/operational_intent_references/$operation_id"  \
-H "Authorization: Bearer ${ACCESS_TOKEN}"  \
-H "Content-Type: application/json"
