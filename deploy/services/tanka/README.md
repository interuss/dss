# Kubernetes deployment via Tanka

This folder contains a set of configuration files to be used by
[tanka](https://tanka.dev/install) to deploy a single DSS instance via
Kubernetes following the procedures found in the [build](../../../build) folder.

## Job cleanup

Job are not automatically removed. It is possible to use tanka to delete unmanaged resources (eg previous jobs) by enabling the [garbage collection](https://tanka.dev/garbage-collection/) feature in your `spec.json` file. Use `tk prune` to cleanup resources not present anymore in the jsonnet.
