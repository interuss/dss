# Tanka library

A set of configuration to be used by [tanka](https://tanka.dev/install) to deploy a single DSS instance via
Kubernetes is provided [there](https://github.com/interuss/dss/tree/master/deploy/services/tanka).

## Requirements

This section hasn't been written yet.

## Usage

This section hasn't been written yet.

## Job cleanup

Jobs are not automatically removed. It is possible to use tanka to delete unmanaged resources (eg previous jobs) by enabling the [garbage collection](https://tanka.dev/garbage-collection/) feature in your `spec.json` file. Use `tk prune` to cleanup resources not present anymore in the jsonnet.
