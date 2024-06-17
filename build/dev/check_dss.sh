#!/usr/bin/env bash

set -eo pipefail

# This script will verify basic functionality of a locally-deployed standalone
# DSS instance using any of the deployment methods described in
# standalone_instance.md.

# Retrieve token from auth server
ACCESS_TOKEN=$(curl --silent \
  "http://localhost:8000/token?grant_type=client_credentials&scope=rid.display_provider&intended_audience=localhost&issuer=localhost&apikey=brutm&aud=localhost" \
  | python extract_json_field.py access_token)

# Retrieve Identification Service Areas currently active on Mauna Loa
echo "DSS response to Mauna Loa ISA query:"
echo "============="
TIMESTAMP_NOW=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
curl --silent -X GET \
  "http://localhost:8082/rid/v2/dss/identification_service_areas?area=19.4763,-155.6043,19.4884,-155.5746,19.4516,-155.5941&earliest_time=${TIMESTAMP_NOW}&latest_time=${TIMESTAMP_NOW}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
echo
echo "============="
echo "See https://tiny.cc/dssapi_rid for a more complete DSS API description."
echo
