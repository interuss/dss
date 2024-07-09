#!/usr/bin/env bash

set -eo pipefail

# Retrieve token from dummy OAuth server
ACCESS_TOKEN=$(curl --silent \
    "http://localhost:8085/token?grant_type=client_credentials&scope=interuss.versioning.read_system_versions&intended_audience=localhost&issuer=localhost&sub=check_scd" \
| python extract_json_field.py 'access_token')

curl --silent -X GET  \
"http://localhost:8082/versions/local.test.identity" \
-H "Authorization: Bearer ${ACCESS_TOKEN}" -H "Content-Type: application/json"

