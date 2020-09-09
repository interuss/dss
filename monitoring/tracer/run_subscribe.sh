#!/usr/bin/env bash

AUTH='--auth=DummyOAuth(https://auth.example.com,uss1)'
DSS='--dss=http://dss.example.com'
AREA='--area=34.5678,-89.1234,34.5689,-89.1245'
LOGS='--output-folder=/config/logs'
BASE_URL='--base-url=https://uss.example.com'
MONITOR='--monitor-rid --monitor-scd'
PORT=5000

TRACER_OPTIONS="$AUTH $DSS $AREA $LOGS $BASE_URL $MONITOR"

docker run --rm -e TRACER_OPTIONS="${TRACER_OPTIONS}" -p ${PORT}:5000 -v `pwd`:/config interuss/dss/tracer gunicorn --workers=2 --bind=0.0.0.0:5000 monitoring.tracer.uss_receiver:webapp
