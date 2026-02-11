# Tanka library

This folder contains a set of configuration to be used with [tanka](https://tanka.dev/) to deploy a single DSS instance in a Kubernetes cluster.

## Requirements

1. A Kubernetes cluster should be running and you should be properly authenticated. Requirements and instructions to create a new Kubernetes cluster can be found here:
    * [AWS](../../../docs/infrastructure/aws.md)
    * [Google](../../../docs/infrastructure/google.md)
    * [Minikube](../../../docs/infrastructure/minikube.md)

2. Create the certificates and apply them to the cluster using the instructions [here](../../../docs/operations/certificates-management.md)
3. Install [Tanka](https://tanka.dev/install)

## Usage

There is a base file in `examples/minikube`. You can directly use it to deploy the service.

Check first that you cluster is using `192.168.49.2` as IP with `minikube ip -p dss-local-cluster`. If that not the case, edit `examples/minikube/spec.json` accordingly.

To apply changes, use the following command:

* `cd examples/minikube`
* `tk apply .`

You may edit `main.jsonnet` as needed should you want to change configuration, or docker images used.

## Job cleanup

Jobs are not automatically removed. It is possible to use tanka to delete unmanaged resources (eg previous jobs) by enabling the [garbage collection](https://tanka.dev/garbage-collection/) feature in your `spec.json` file. Use `tk prune` to cleanup resources not present anymore in the jsonnet.
