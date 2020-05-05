#!/bin/bash

# This script will verify basic functionality of a locally-deployed standalone
# DSS instance using any of the deployment methods described in
# standalone_instance.md.

jq --version > /dev/null
if [ $? -ne 0 ]; then
  echo "This script requires the jq utility.  On Debian Linux, install with"
  echo "  sudo apt-get install jq"
  echo "With homebrew, install with"
  echo "  brew install jq"
  exit 1
fi

# Retrieve token from dummy OAuth server
export ACCESS_TOKEN_READ=`curl --silent -X POST \
  "http://localhost:8085/token?grant_type=client_credentials&scope=dss.read.identification_service_areas&intended_audience=localhost&issuer=localhost" \
  | jq -r '.access_token'`

export ACCESS_TOKEN_WRITE=`curl --silent -X POST \
  "http://localhost:8085/token?grant_type=client_credentials&scope=dss.write.identification_service_areas&intended_audience=localhost&issuer=localhost" \
  | jq -r '.access_token'`

echo "DSS response to PUT Subscriptions query:"
echo "============="
curl --silent -X PUT \
  "http://localhost:8082/dss/v1/subscriptions/b76c1049-94e3-47e5-900d-94e4004a7188" \
  -H "Authorization: Bearer ${ACCESS_TOKEN_WRITE}" \
  -H "Content-Type: application/json" \
  -d '{
  "extents": {
    "volume": {
      "outline_circle": {
        "type": "Feature",
        "geometry": {
          "type": "Point",
          "coordinates": {
            "type": "Point",
            "coordinates": [
              -122.106325,
              47.660898
            ]
          }
        },
        "properties": {
          "radius": {
            "value": 300.183,
            "units": "M"
          }
        }
      },
      "altitude_lower": {
        "value": 0,
        "reference": "W84",
        "units": "M"
      },
      "altitude_upper": {
        "value": 3000,
        "reference": "W84",
        "units": "M"
      }
    },
    "time_start": {
      "value": "1985-04-12T23:20:50.52Z",
      "format": "RFC3339"
    },
    "time_end": {
      "value": "2100-04-12T23:20:50.52Z",
      "format": "RFC3339"
    }
  },
  "old_version": 0,
  "uss_base_url": "https://exampleuss.com/utm",
  "notify_for_operations": true,
  "notify_for_constraints": false
}'
echo
echo "============="
echo

exit 0

echo "DSS response to GET Subscription query:"
echo "============="
curl --silent -X GET \
  "http://localhost:8082/dss/v1/subscriptions/b76c1049-94e3-47e5-900d-94e4004a7188" \
  -H "Authorization: Bearer ${ACCESS_TOKEN_READ}"
echo
echo "============="
echo

echo "DSS response to DELETE Subscription query:"
echo "============="
curl --silent -X DELETE \
  "http://localhost:8082/dss/v1/subscriptions/b76c1049-94e3-47e5-900d-94e4004a7188" \
  -H "Authorization: Bearer ${ACCESS_TOKEN_WRITE}"
echo
echo "============="
echo

echo "DSS response to search Subscriptions query:"
echo "============="
curl --silent -X POST \
  "http://localhost:8082/dss/v1/subscriptions/query" \
  -H "Authorization: Bearer ${ACCESS_TOKEN_READ}" \
  -H "Content-Type: application/json" \
  -d '{
  "area_of_interest": {

  }
}'
echo
echo "============="
echo
