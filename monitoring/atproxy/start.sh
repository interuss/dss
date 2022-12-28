#!/usr/bin/env bash

# This script is intended to be called from within a Docker container running
# atproxy via the interuss/monitoring image.  In that context, this script is
# the entrypoint into the atproxy server.

# Ensure atproxy is the working directory
OS=$(uname)
if [[ $OS == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi
cd "${BASEDIR}" || exit 1

# Use atproxy's health check
cp health_check.sh /app

# Start atproxy server on port 5000
gunicorn \
    --preload \
    --workers=4 \
    --timeout 60 \
    --bind=0.0.0.0:5000 \
    monitoring.atproxy.app:webapp
