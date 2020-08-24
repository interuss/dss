#!/usr/bin/env sh

COMMIT=$(git rev-parse --short HEAD)

if  test -n "$(git status -s)"; then
    echo "${COMMIT}-dirty"
elif test -n "$(git cherry 2> /dev/null)"; then
    echo "${COMMIT}-localcommit"
#TODO: Handle the case where the current branch does not exist remotely
else
    echo "${COMMIT}"
fi
