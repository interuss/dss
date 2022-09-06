#!/usr/bin/env sh

# This script is intended to be called from within a Docker container running
# atproxy via the interuss/monitoring image to determine the health status of
# the container.

curl --fail http://localhost:5000/ || exit 1
