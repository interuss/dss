#!/usr/bin/env bash

# This script prints the organization of the upstream repository using the remote origin url.
# It expects a github.com remote defined as origin and the following url formats:
# 1. git@github.com:interuss/dss.git
# 2. git@github.com/interuss/dss.git
# 3. https://github.com/interuss/dss.git

UPSTREAM_REPO=$(git remote get-url origin)
# Replace `:` by `/` to handle git@github.com:interuss/dss.git remote reference.
UPSTREAM_REPO=${UPSTREAM_REPO//:/\/}
# Remove hostname part
UPSTREAM_OWNER=$(dirname ${UPSTREAM_REPO#*github.com/*})

echo $UPSTREAM_OWNER
