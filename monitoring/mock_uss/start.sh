#!/usr/bin/env bash

# This script is intended to be called from within a Docker container running
# mock_uss via the interuss/monitoring image.  In that context, this script is
# the entrypoint into the mock_uss server.

# Ensure mock_uss is the working directory
OS=$(uname)
if [[ $OS == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi
cd "${BASEDIR}" || exit 1

# Use mock_uss's health check
cp health_check.sh /app

# Start mock_uss server on port 5000
gunicorn \
    --preload \
    --workers=4 \
    --bind=0.0.0.0:5000 \
    monitoring.mock_uss:webapp
