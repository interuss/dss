#!/usr/bin/env sh

OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi
cd "${BASEDIR}"

# We only enable -o pipefail after having verified that
# the command line argument satisfies format requirements.
version=$(echo "$1" | grep -E 'v[0-9]+\.[0-9]+\.[0-9]+')

set -e

branch=$(git rev-parse --abbrev-ref HEAD)

if test "${branch}" != "master"; then
  echo "releases are only supported on master" && false
fi

if test -z "${version}"; then
  echo "${1} does not match v[0-9]+\.[0-9]+\.[0-9]+" && false
fi

UPSTREAM_ORG=$(./upstream_organization.sh)

git tag -a "${UPSTREAM_ORG}/dss/${version}"
git push tag "${UPSTREAM_ORG}/dss/${version}"
