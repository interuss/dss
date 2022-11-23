#!/usr/bin/env sh

# We only enable -o pipefail after having verified that
# the command line argument satisfies format requirements.
# Semantic versioning regex (suffixed below) from:
# https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
tag_regex='^[^/]+/[^/]+/v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$'
tag=$(echo "$1" | grep -E "${tag_regex}")

set -e

if test -z "${tag}"; then
  echo "requested tag \"${1}\" does not match expected tag format [owner]/[component]/[semantic version] using the pattern ${tag_regex}" && false
fi

branch=$(git rev-parse --abbrev-ref HEAD)

if test "${branch}" != "master"; then
  echo "releases are only supported on master branch (currently on ${branch})" && false
fi

if test -n "$(git status -s)"; then
  echo "releases are only supported in a clean git workspace" && false
fi

git tag -a "${tag}"
git push origin "${tag}"
