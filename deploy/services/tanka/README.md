# Kubernetes deployment via Tanka

This folder contains a set of configuration files to be used by
[tanka](https://tanka.dev/install) to deploy a single DSS instance via
Kubernetes following the procedures found in the [build](../../../build) folder.

## Job cleanup

Job can build-up, if you don't want to remove them manully, enable the [inject labels](https://tanka.dev/garbage-collection/) feature in your `spec.json` file and use `tk prune` to do some cleanup.
