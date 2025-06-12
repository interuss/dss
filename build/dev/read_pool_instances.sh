#!/usr/bin/env bash

set -eo pipefail

# Retrieve token from dummy OAuth server
ACCESS_TOKEN=$(curl --silent \
    "http://localhost:8085/token?grant_type=client_credentials&scope=interuss.pool_status.read&intended_audience=localhost&issuer=localhost&sub=manual_tester" \
| python extract_json_field.py 'access_token')

curl --silent -X GET  \
"http://localhost:8082/aux/v1/pool/dss_instances" \
-H "Authorization: Bearer ${ACCESS_TOKEN}" -H "Content-Type: application/json"
