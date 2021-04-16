#!/usr/bin/env bash

AUTH='--auth=DummyOAuth(http://host.docker.internal:8085/token,uss1)'
DSS='--dss=http://host.docker.internal:8082'
AREA='--area=34.1234,-123.4567,34.4567,-123.1234'
LOGS='--output-folder=/logs'
BASE_URL='--base-url=http://host.docker.internal:5000'
KML_SERVER='--kml-server=https://example.com/kmlgeneration'
KML_FOLDER='--kml-folder=test/localmock'
MONITOR='--monitor-rid --monitor-scd'
PORT=5000

TRACER_OPTIONS="$AUTH $DSS $AREA $LOGS $BASE_URL $KML_SERVER $KML_FOLDER $MONITOR"

echo Reminder: must be run from root repo folder

docker build \
    -f monitoring/tracer/Dockerfile \
    -t interuss/dss/tracer \
    --build-arg version=`scripts/git/commit.sh` \
    monitoring

docker run --name tracer_subscribe \
  --rm \
  -e TRACER_OPTIONS="${TRACER_OPTIONS}" \
  -p ${PORT}:5000 \
  -v /Users/pelletierb/Documents/test/localmock:/logs \
  interuss/dss/tracer \
  gunicorn \
    --preload \
    --workers=2 \
    --bind=0.0.0.0:5000 \
    monitoring.tracer.uss_receiver:webapp
