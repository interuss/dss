#!/usr/bin/env sh

# This script prints the current version of a component in the repository based on the tags
# of the upstream repository (remote origin) matching the following convention:
# remote_origin/component/version. Examples of values:
# - remote_origin: interuss
# - component: rid, scd, aux, uss_qualifier
# - version: v3.0.1[-hash][-dirty]
#    - [-hash] (example: -8a493ef8 ) is added when commits have been added to the latest version tagged.
#    - [-dirty] (example: -dirty) when the workspace is not clean.
# Only versions without [-hash] and without [-dirty] shall be released.

OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi
cd "${BASEDIR}"

if [[ $# == 0 ]]; then
  echo "Usage: $0 <COMPONENT> [--long]"
  echo "Print the component's version number. (ie v0.0.1)"
  echo "[--long]: Print the component's version using the long format including the upstream organization (ie interuss/scd/v0.0.1)."
  exit 1
fi

COMPONENT=${1:?"Component must be provided as environment variable. (example: scd, rid, aux, uss_qualifier)"}

RELEASE_FORMAT=false
if [[ $2 == "--long" ]]; then
  RELEASE_FORMAT=true
fi

UPSTREAM_ORG=$(./upstream_organization.sh)

# Look for the last tag of the component
LAST_VERSION_TAG=$(git describe --abbrev=1 --tags --match="${UPSTREAM_ORG}/${COMPONENT}/*" 2> /dev/null)
#echo "LAST_VERSION_TAG: $LAST_VERSION_TAG"

# Store in LAST_VERSION the version of the tag (ie v0.0.1)
LAST_VERSION=${LAST_VERSION_TAG##*/}
#echo "LAST_VERSION: $LAST_VERSION"

# If no version was found, use default v0.0.0.
if [[ -z "$LAST_VERSION" ]]; then
  LAST_VERSION="v0.0.0"
fi

# Current commit
COMMIT=$(git rev-parse --short HEAD)

# Check if there are some commits on top of the tag by checking if an abbrev part is present.
if [[ "$LAST_VERSION" == *"-"* ]]; then
  # Remove abbrev part
  LAST_VERSION=${LAST_VERSION%%-*}
  # Append the commit hash
  LAST_VERSION=${LAST_VERSION}-${COMMIT}
fi

# Set the dirty flag if the workspace is not clean.
if  test -n "$(git status -s)"; then
    DIRTY="-dirty"
fi

if [[ "$RELEASE_FORMAT" == "true" ]]; then
  echo ${UPSTREAM_ORG}/${COMPONENT}/${LAST_VERSION}${DIRTY}
else
  echo ${LAST_VERSION}${DIRTY}
fi
