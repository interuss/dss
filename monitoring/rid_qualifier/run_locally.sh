
#!/usr/bin/env bash

AUTH='--auth=NoAuth()'
# NB: A prerequisite to run this command locally is to have a running local DSS instance session via '/build/dev/run_locally.sh', for more information see https://github.com/interuss/dss/blob/master/build/dev/standalone_instance.md

LOCALE='--locale=che'

INJECTION_URL='--injection_base_url=http://localhost:8070/sp/uss1'


RID_QUALIFIER_OPTIONS="$AUTH $LOCALE $INJECTION_URL"

echo Reminder: must be run from root repo folder

docker build \
    -f monitoring/rid_qualifier/Dockerfile \
    -t interuss/dss/rid_qualifier \
    --build-arg version=`scripts/git/commit.sh` \
    monitoring

docker run --name rid_qualifier \
  --rm \
  -e RID_QUALIFIER_OPTIONS="${RID_QUALIFIER_OPTIONS}" \
  interuss/dss/rid_qualifier \
  python rid_qualifier_entry.py $RID_QUALIFIER_OPTIONS