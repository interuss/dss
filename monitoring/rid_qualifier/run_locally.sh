
#!/usr/bin/env bash

AUTH='--auth=NoAuth()'
# NB: A prerequisite to run this command locally is to have a running instance of the rid_qualifier/mock (/monitoring/rid_qualifier/mock/run_locally.sh)

LOCALE='--locale=che'

INJECTION_URL='--injection_base_url=http://host.docker.internal:8070/sp/uss1'

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
