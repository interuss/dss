#!/usr/bin/env bash

set -eo pipefail

# Retrieve token from dummy OAuth server
ACCESS_TOKEN=$(curl --silent \
    "http://localhost:8085/token?grant_type=client_credentials&scope=interuss.pool_status.read&intended_audience=localhost&issuer=localhost&sub=check_scd" \
| python extract_json_field.py 'access_token')

curl --silent -X GET  \
"http://localhost:8082/aux/v1/configuration/scd_lock_mode" \
-H "Authorization: Bearer ${ACCESS_TOKEN}" -H "Content-Type: application/json"
