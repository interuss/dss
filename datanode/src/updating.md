# Updating version info

When making changes, version information needs to be updated in several places.

## src/storage_api.py

Comment out the current `VERSION = ` line and add a new line below.

## docker/Dockerfile_storageapi

Change the version in the `docker tag ` comment line to match to the new version.
