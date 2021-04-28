
#!/usr/bin/env bash

AUTH='--auth=DummyOAuth(http://host.docker.internal:8085/token,uss1)'
DSS='--dss=http://host.docker.internal:8082'
PORT=5000

RID_QUALIFIER_OPTIONS="$AUTH $DSS"

echo Reminder: must be run from root repo folder

docker build \
    -f monitoring/rid_qualifier/Dockerfile \
    -t interuss/dss/rid_qualifier \
    --build-arg version=`scripts/git/commit.sh` \
    monitoring

docker run --name rid_qualifier \
  --rm \
  -v "$(pwd)"/test_definitions:/test_definitions \
  interuss/dss/rid_qualifier \
    