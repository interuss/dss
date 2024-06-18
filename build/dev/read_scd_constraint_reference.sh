#!/usr/bin/env bash

set -eo pipefail

# Get named SCD constraint.

constraint_id=$1
[[ -z "$constraint_id" ]] && { echo "Error: Constraint ID not provided."; exit 1; }

# Retrieve token from auth server
ACCESS_TOKEN=$(curl --silent \
    "http://localhost:8000/token?grant_type=client_credentials&scope=utm.constraint_processing&intended_audience=localhost&issuer=localhost&sub=check_scd&apikey=brutm&aud=localhost" \
| python extract_json_field.py 'access_token')

curl --silent -X GET  "http://localhost:8082/dss/v1/constraint_references/$constraint_id"  \
-H "Authorization: Bearer ${ACCESS_TOKEN}"  \
-H "Content-Type: application/json"
