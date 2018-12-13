# Updating version information

## Version format

Versions follow the format `BRANCH.MAJOR.MINOR.BUILD`

* **BRANCH** is the specific version for this branch, or numeric for the
mainline branch, changing only on new branches or fundamental changes to the
mainline branch.
* **MAJOR** changes when there is major functionality change that either
directly or eventually deprecates functionality in previous major versions.
* **MINOR** changes when there is a non-breaking API change (new fields or
methods) and resets to 0 when a MAJOR change takes place.
* **BUILD** continuously increments with every release, regardles of other
version number changes.

## What to update

When making changes, version information needs to be updated in multiple places.

### datanode/src/storage_api.py

Comment out the current `VERSION = ` line and add a new line below.

### datanode/docker/Dockerfile_storageapi

Change the version in the `docker tag ` comment line to match to the new version.

### assets/swaggerhub_api.yaml

Change the version tag to match the new version.

## When making a new branch

A new branch requires changes in additional places as well.

### datanode/docker/docker-compose.yaml

Change the storage_api image to target the new branch.

### datanode/docker/docker-compose_storageapitest.yaml

Change the storage_api image to target the new branch.

### datanode/docker/node.sh

Change the deployment file download locations.

### datanode/docker/Dockerfile_storageapi

Also change the `docker image build` and `docker run` comment lines to target
the new branch.

### datanode/docker/README.md

Change anywhere `interussplatform/storage_api` is referenced to also include the
branch.
