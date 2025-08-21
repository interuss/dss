# Certificates management

## Introduction

The `dss-certs.py` helps you manage the set of certificates used for your DSS deployment.

Should this DSS beeing part of a pool, the script also provide some helpers to manage the set of CA certificates in the pool.

To run the script, just run `./dss-certs.py`. The python script don't require any dependencies, just a recent version of python 3.

## Quick start guide

### Single DSS instance in minikube`

* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default init`
* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default apply`

### Pool of 3 DSS instances in minikube, in namespace `default`, `ns2` and `ns3`

* Creation of the 3 DSS instances certificates
* `./dss-certs.py --name dss-instance-1 --cluster-context dss-local-cluster --namespace default init`
* `./dss-certs.py --name dss-instance-2 --cluster-context dss-local-cluster --namespace ns2 init`
* `./dss-certs.py --name dss-instance-1 --cluster-context dss-local-cluster --namespace ns3 init`
* Copy instance 2 and 3 CA certificates to the instance 1
* `./dss-certs.py --name dss-instance-2 --cluster-context dss-local-cluster --namespace ns2 get-ca | ./dss-certs.py --name dss-instance-1 --cluster-context dss-local-cluster --namespace default add-pool-ca`
* `./dss-certs.py --name dss-instance-3 --cluster-context dss-local-cluster --namespace ns3 get-ca | ./dss-certs.py --name dss-instance-1 --cluster-context dss-local-cluster --namespace default add-pool-ca`
* Reuse instance compiled 1 CA and copy it to instance 2 and 3.
* `./dss-certs.py --name dss-instance-1 --cluster-context dss-local-cluster --namespace default get-pool-ca | ./dss-certs.py --name dss-instance-2 --cluster-context dss-local-cluster --namespace ns2 add-pool-ca`
* `./dss-certs.py --name dss-instance-1 --cluster-context dss-local-cluster --namespace default get-pool-ca | ./dss-certs.py --name dss-instance-3 --cluster-context dss-local-cluster --namespace ns3 add-pool-ca`
* Application of certificates in respective clusters
* `./dss-certs.py --name dss-instance-1 --cluster-context dss-local-cluster --namespace default apply`
* `./dss-certs.py --name dss-instance-2 --cluster-context dss-local-cluster --namespace ns2 apply`
* `./dss-certs.py --name dss-instance-3 --cluster-context dss-local-cluster --namespace ns3 apply`

## Operations

### Common parameters

#### `--name`

The name of your DSS instance, that should identify it in a unique way. Used as main identifier for the set of certificates and in certificates.

Example: `dss-west-1`

#### `--organization`

The name of the organization managing the DSS Instance. Used in certificates generation. The combination of (name, organization) shall be unique in a cluster.

Example: `Interuss`

#### `--cluster-context`

The kubernetes context the script should use.

Example: `dss-local-cluster`

#### `--namespace`

The kubernetes namespace to use.

Example: `default`

#### `--nodes-count`

The number of yugabyte nodes of your DSS instance. Default to `3`.

### `init`

Initializes the certificates for a new DSS instance including a CA, a client certificate and a certificate for each yugabyte node.

### `apply`

Apply the current set of certificates to the kubernetes cluster. Shall be ran after each modification of the certificates, like addition / removal of CA in the pool, new `nodes-count` parameter.

### `regenerate-nodes`

Generate missing nodes certificates. Useful if you want to add new nodes in your DSS Instance. Don't forget to set the `nodes-count` parameters.

### `add-pool-ca`

Add a CA certificate(s) of another(s) DSS Instance to the set of trusted certificates.
Existing certificates are not added again.

You can set the file with certificate(s) with `--ca-file` or use stdin.

Don't forget to use the `apply` command to update certificate on your kubernetes cluster.

Examples:

* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default add-pool-ca < /tmp/new-dss-ca`
* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default --ca-file /tmp/new-dss-ca add-pool-ca`
* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default get-pool-ca | ./dss-certs.py --name test2 --cluster-context dss-local-cluster --namespace namespace2 add-pool-ca`

### `remove-pool-ca`

Remove CA certificate(s) of DSS Instance(s) from the set of trusted certificates.
Unknown certificates are not removed again.

You can set the file with certificate(s) with `--ca-file`, use stdin or use `--ca-serial` to specify the serial / name of the certificate you want to remove.

Don't forget to use the `apply` command to update certificate on your kubernetes cluster.

Example:

* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default remove-pool-ca < /tmp/old-dss-ca`
* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default --ca-file /tmp/old-dss-ca remove-pool-ca`
* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default remove-pool-ca --ca-serial="SN=830ECFB0, O=generic-dss-organization, CN=CA.test"`
* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default remove-pool-ca --ca-serial="830ECFB0`
* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default remove-pool-ca --ca-serial="46548B7CC9699A7CFA54FF8FA85A619E830ECFB0`

### `list-pool-ca`

List the set of accepted CA certificates.

Also display a 'hash' of CA serial, that you may use to compare other DSS Instances list of CA certificates easily.

### `get-pool-ca`

Return all CA certificate in the current pool.

Can be used for debugging or to synchronize the set of CA certificates in a pool with others USS.

### `get-ca`

Return your own CA certificate .

Display the compiled CA certificate. Can be used for debugging or to synchronize the set of CA certificates in a pool with others USS.

### `destroy`

Destroy a certificate set. Be careful, there are no way to undo the command.
