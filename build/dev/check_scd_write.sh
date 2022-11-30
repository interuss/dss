#!/usr/bin/env bash

set -eo pipefail

# This script is to perform simple write operations on SCD database for testing.

# Retrieve token from dummy OAuth server
ACCESS_TOKEN=$(curl --silent \
    "http://localhost:8085/token?grant_type=client_credentials&scope=utm.strategic_coordination%20utm.constraint_management&intended_audience=localhost&issuer=localhost&sub=check_scd" \
| python extract_json_field.py 'access_token')


echo "DSS response to [SCD] PUT subscription query:"
echo "============="
TIMESTAMP_NOW=$(python -c 'from datetime import datetime; print((datetime.utcnow()).isoformat() + "Z")')
TIMESTAMP_LATER=$(python -c 'from datetime import datetime, timedelta; print((datetime.utcnow() + timedelta(minutes=5)).isoformat() + "Z")')


curl --silent -X PUT  \
"http://localhost:8082/dss/v1/subscriptions/00000158-9aba-4026-bbef-bad6cde80000" \
-H "Authorization: Bearer ${ACCESS_TOKEN}" -H "Content-Type: application/json"  \
-d '{
       "notify_for_operational_intents": true,
       "notify_for_constraints": false,
       "uss_base_url": "https://example.com/foo",
       "extents": {
           "volume": {
               "altitude_upper": {
                   "units": "M",
                   "reference": "W84",
                   "value": 300
                },
                "altitude_lower": {
                    "units": "M",
                    "reference": "W84",
                    "value": 0
                },
                "outline_circle": {
                    "radius": {
                        "units": "M",
                        "value": 100
                    },
                    "center": {
                        "lat": 22.910168434185902,
                        "lng": 56
                    }
            }},
            "time_end": {
                "value": "'"$TIMESTAMP_LATER"'",
                "format": "RFC3339"
            }
}}'
echo


echo "DSS response to [SCD] PUT constraint reference query:"
echo "============="

curl --silent -X PUT  "http://localhost:8082/dss/v1/constraint_references/00000159-9aba-4026-bbef-bad6cde80000"  \
-H "Authorization: Bearer ${ACCESS_TOKEN}"  \
-H "Content-Type: application/json"  \
-d '{
    "uss_base_url": "https://example.com/con1/uss",
    "extents": [
        {
            "volume": {
                "altitude_upper": {
                    "units": "M",
                    "reference": "W84",
                    "value": 120
                },
                "altitude_lower": {
                    "units": "M",
                    "reference": "W84",
                    "value": 0
                },
                "outline_circle": {
                    "radius": {
                        "units": "M",
                        "value": 50
                    },
                    "center": {
                        "lat": -12.00001,
                        "lng": 33.99999
                    }
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
    ],
    "old_version": 0
}'
echo


echo "DSS response to [SCD] PUT operational intent reference query:"
echo "=========="

curl --silent -X PUT  "http://localhost:8082/dss/v1/operational_intent_references/0000015a-9aba-4026-bbef-bad6cde80000"  \
 -H "Authorization: Bearer ${ACCESS_TOKEN}"  \
 -H "Content-Type: application/json"  \
 -d '{
    "state": "Accepted",
    "uss_base_url": "https://localhost:8080",
    "extents": [
        {
            "volume": {
                "altitude_upper": {
                    "units": "M",
                    "reference": "W84",
                    "value": 114.09273234903256
                },
                "altitude_lower": {
                    "units": "M",
                    "reference": "W84",
                    "value": -7.827263749527558
                },
                "outline_circle": {
                    "radius": {
                        "units": "M",
                        "value": 300.183
                    },
                    "center": {
                        "lat": 32.59415886402617,
                        "lng": -117.1035655786647
                    }
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
        {
            "volume": {
                "altitude_upper": {
                    "units": "M",
                    "reference": "W84",
                    "value": 114.09273234903256
                },
                "altitude_lower": {
                    "units": "M",
                    "reference": "W84",
                    "value": 98.85273283671255
                },
                "outline_polygon": {
                    "vertices": [
                        {
                            "lat": 32.629403155814224,
                            "lng": -117.07235209637365
                        },
                        {
                            "lat": 32.62913940035541,
                            "lng": -117.07227677467183
                        },
                        {
                            "lat": 32.61019036953776,
                            "lng": -117.15071909262814
                        },
                        {
                            "lat": 32.61045408389095,
                            "lng": -117.15079458449262
                        }
                    ]
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
        {
            "volume": {
                "altitude_upper": {
                    "units": "M",
                    "reference": "W84",
                    "value": 114.09273234903256
                },
                "altitude_lower": {
                    "units": "M",
                    "reference": "W84",
                    "value": 98.85273283671255
                },
                "outline_polygon": {
                    "vertices": [
                        {
                            "lat": 32.63060966167485,
                            "lng": -117.09038826289287
                        },
                        {
                            "lat": 32.630365389882776,
                            "lng": -117.09026347832622
                        },
                        {
                            "lat": 32.60141106631151,
                            "lng": -117.15717622635013
                        },
                        {
                            "lat": 32.601655284082824,
                            "lng": -117.15730113575759
                        }
                    ]
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
        {
            "volume": {
                "altitude_upper": {
                    "units": "M",
                    "reference": "W84",
                    "value": 114.09273234903256
                },
                "altitude_lower": {
                    "units": "M",
                    "reference": "W84",
                    "value": -7.827263749527558
                },
                "outline_polygon": {
                    "vertices": [
                        {
                            "lat": 32.63712150700819,
                            "lng": -117.1102585196023
                        },
                        {
                            "lat": 32.636969253183054,
                            "lng": -117.11003035210928
                        },
                        {
                            "lat": 32.58489223521358,
                            "lng": -117.15104910529122
                        },
                        {
                            "lat": 32.5850444653223,
                            "lng": -117.15127729146461
                        }
                    ]
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
        {
            "volume": {
                "altitude_upper": {
                    "units": "M",
                    "reference": "W84",
                    "value": 7.412735762792458
                },
                "altitude_lower": {
                    "units": "M",
                    "reference": "W84",
                    "value": -20.01926335938357
                },
                "outline_polygon": {
                    "vertices": [
                        {
                            "lat": 32.58037888894311,
                            "lng": -117.13297510795468
                        },
                        {
                            "lat": 32.58037888894311,
                            "lng": -117.13324941122436
                        },
                        {
                            "lat": 32.635427303056666,
                            "lng": -117.13324941122436
                        },
                        {
                            "lat": 32.635427303056666,
                            "lng": -117.13297510795468
                        }
                    ]
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
    ],
    "new_subscription": {
        "uss_base_url": "https://localhost:8080",
        "notify_for_constraints": true
    },
    "key": []
}'
echo
