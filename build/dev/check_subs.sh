#!/usr/bin/env bash

set -eo pipefail

# This script will verify basic functionality of a locally-deployed standalone
# DSS instance using any of the deployment method described in
# standalone_instance.md.

if jq --version > /dev/null; then
  echo "This script requires the jq utility.  On Debian Linux, install with"
  echo "  sudo apt-get install jq"
  echo "With homebrew, install with"
  echo "  brew install jq"
  exit 1
fi

# Retrieve token from dummy OAuth server
ACCESS_TOKEN=$(curl --silent -X POST \
  "http://localhost:8085/token?grant_type=client_credentials&scope=utm.strategic_coordination&intended_audience=localhost&issuer=localhost" \
  | jq -r '.access_token')

echo "DSS response to [SCD] PUT Subscriptions query:"
echo "============="
TIMESTAMP_NOW=$(python -c 'from datetime import datetime; print((datetime.utcnow()).isoformat() + "Z")')
TIMESTAMP_LATER=$(python -c 'from datetime import datetime, timedelta; print((datetime.utcnow() + timedelta(minutes=5)).isoformat() + "Z")')
curl --silent -X PUT \
  "http://localhost:8082/dss/v1/subscriptions/b76c1049-94e3-47e5-900d-94e4004a7188" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
  "extents": {
    "volume": {
      "outline_circle": {
        "center": {
          "lng": -122.106325,
          "lat": 47.660898
        },
        "radius": {
          "value": 300.183,
          "units": "M"
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
      "value": "'"$TIMESTAMP_NOW"'",
      "format": "RFC3339"
    },
    "time_end": {
      "value": "'"$TIMESTAMP_LATER"'",
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

echo "DSS response to [SCD] GET Subscription query:"
echo "============="
curl --silent -X GET \
  "http://localhost:8082/dss/v1/subscriptions/b76c1049-94e3-47e5-900d-94e4004a7188" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
echo
echo "============="
echo

echo "DSS response to [SCD] DELETE Subscription query:"
echo "============="
curl --silent -X DELETE \
  "http://localhost:8082/dss/v1/subscriptions/b76c1049-94e3-47e5-900d-94e4004a7188" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
echo
echo "============="
echo

echo "DSS response to search [SCD] Subscriptions query:"
echo "============="
curl --silent -X POST \
  "http://localhost:8082/dss/v1/subscriptions/query" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
  "area_of_interest": {
    "volume": {
      "outline_circle": {
        "center": {
          "lng": -122.106325,
          "lat": 47.660898
        },
        "radius": {
          "value": 300.183,
          "units": "M"
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
      "value": "'"$TIMESTAMP_NOW"'",
      "format": "RFC3339"
    },
    "time_end": {
      "value": "'"$TIMESTAMP_LATER"'",
      "format": "RFC3339"
    }
  }
}'
echo
echo "============="
echo
