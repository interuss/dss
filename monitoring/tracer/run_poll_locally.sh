#!/usr/bin/env bash

AUTH='--auth=DummyOAuth(http://host.docker.internal:8085/token,uss1)'
DSS='--dss=http://host.docker.internal:8082'
AREA='--area=34.1234,-123.4567,34.4567,-123.1234'
LOGS='--output-folder=/logs'
KML_SERVER='--kml-server=https://example.com/kmlgeneration'
KML_FOLDER='--kml-folder=test/localmock'
MONITOR='--rid-isa-poll-interval=15 --scd-operation-poll-interval=15 --scd-constraint-poll-interval=15'
PORT=5000

TRACER_OPTIONS="$AUTH $DSS $AREA $LOGS $KML_SERVER $KML_FOLDER $MONITOR"

echo Reminder: must be run from root repo folder

docker build \
    -f monitoring/tracer/Dockerfile \
    -t interuss/dss/tracer \
    --build-arg version=`scripts/git/commit.sh` \
    monitoring

docker run --name tracer_poll \
  --rm \
  -e TRACER_OPTIONS="${TRACER_OPTIONS}" \
  -v /Users/pelletierb/Documents/test/localmock:/logs \
  interuss/dss/tracer \
  python tracer_poll.py $TRACER_OPTIONS
