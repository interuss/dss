#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

docker build -t terraform-variables:latest .
docker run \
    -v "${SCRIPT_DIR}/":/app/utils:rw \
    -v "${SCRIPT_DIR}/../dependencies":/app/examples:rw \
    -v "${SCRIPT_DIR}/../modules":/app/modules:rw \
    -w /app/utils \
    terraform-variables \
    ls ../ && python variables.py