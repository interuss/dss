#!/bin/sh

COMMIT=$(git rev-parse --short HEAD)

if [[ -n $(git status -s) ]] || [[ -n $(git cherry 2> /dev/null) ]]; then 
    echo ${COMMIT}-dirty
else 
    echo ${COMMIT}
fi