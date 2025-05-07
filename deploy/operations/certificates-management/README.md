# Certificates management

## Introduction

The `dss-certs.py` helps you manage the set of certificates used for your DSS deployment.

Should this DSS beeing part of a pool, the script also provide some helpers to manage the set of CA certificates in the pool.

To run the script, just run `./dss-certs.py`. The python script don't require any dependencies, just a recent version of python 3.

## Quick start guide

### Simple local cluster in minikube`

* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default init`
* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default apply`

### Simple pool of 3 local cluster in minikube, in namespace `default`, `ns2` and `ns3`

* Creation of the 3 cluster's certificates
* `./dss-certs.py --name localpool --cluster-context dss-local-cluster --namespace default init`
* `./dss-certs.py --name localpool2 --cluster-context dss-local-cluster --namespace ns2 init`
* `./dss-certs.py --name localpool3 --cluster-context dss-local-cluster --namespace ns3 init`
* Copy of cluster 2 and 3 CA to the base cluster
* `./dss-certs.py --name localpool2 --cluster-context dss-local-cluster --namespace ns2 get-ca | ./dss-certs.py --name localpool --cluster-context dss-local-cluster --namespace default add-pool-ca`
* `./dss-certs.py --name localpool3 --cluster-context dss-local-cluster --namespace ns3 get-ca | ./dss-certs.py --name localpool --cluster-context dss-local-cluster --namespace default add-pool-ca`
* Copy of base cluster's CA pool to cluster 2 and 3's CA pool
* `./dss-certs.py --name localpool --cluster-context dss-local-cluster --namespace default get-pool-ca | ./dss-certs.py --name localpool2 --cluster-context dss-local-cluster --namespace ns2 add-pool-ca`
* `./dss-certs.py --name localpool --cluster-context dss-local-cluster --namespace default get-pool-ca | ./dss-certs.py --name localpool3 --cluster-context dss-local-cluster --namespace ns3 add-pool-ca`
* Application of certificates in respective clusters
* `./dss-certs.py --name localpool --cluster-context dss-local-cluster --namespace default apply`
* `./dss-certs.py --name localpool2 --cluster-context dss-local-cluster --namespace ns2 apply`
* `./dss-certs.py --name localpool3 --cluster-context dss-local-cluster --namespace ns3 apply`

## Operations

### Common parameters

#### `--name`

The name of your cluster, that should identify it in a unique way. Used as main identifier for the set of certificates and in certificates but may be different inside a DSS pool.

Example: `dss-west-1`

#### `--organization`

The name or your organization. Used in certificates generation. The combination of (name, organization) shall be unique in a cluster.

Example: `interuss`

#### `--cluster-context`

The kubernetes context the script should use.

Example: `dss-local-cluster`

#### `--namespace`

The kubernetes namespace to use.

Example: `default`

#### `--nodes-count`

The number of yugabyte nodes you have. Default to `3`.

### `init`

Create a new set of certificates, with a CA, a client certificate and a certificate for each yugabyte node.

### `apply`

Apply the current set of certificate to the kubernetes cluster. Shall be ran after each modification of the certificates, like addition / removal of CA in the pool, new `nodes-count` parameter.

### `regenerate-nodes`

Generate missing nodes certificates. Useful if you want to add new nodes in your cluster. Don't forget to set the `nodes-count` parameters.

### `add-pool-ca`

Add the CA certificate(s) of another(s) USS in the pool of trusted certificates.
Existing certificates are not added again, so you may simply use the output of `get-pool-ca` from another USS.

You can set the file with certificate(s) with `--ca-file` or use stdin.

Don't forget to use the `apply` command to update certificate on your kubernetes cluster.

Examples:

* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default add-pool-ca < /tmp/new-dss-ca`
* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default --ca-file /tmp/new-dss-ca add-pool-ca`
* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default get-pool-ca | ./dss-certs.py --name test2 --cluster-context dss-local-cluster --namespace namespace2 add-pool-ca`

### `remove-pool-ca`

Remove the CA certificate(s) of another(s) USS in the pool of trusted certificates.
Unknown certificates are not remove again.

You can set the file with certificate(s) with `--ca-file`, use stdin or use `--ca-serial` to specify the serial / name of the certificate you want to remove.

Don't forget to use the `apply` command to update certificate on your kubernetes cluster.

Example:

* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default remove-pool-ca < /tmp/old-dss-ca`
* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default --ca-file /tmp/old-dss-ca remove-pool-ca`
* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default remove-pool-ca --ca-serial="SN=830ECFB0, O=generic-dss-organization, CN=CA.test"`
* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default remove-pool-ca --ca-serial="830ECFB0`
* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default remove-pool-ca --ca-serial="46548B7CC9699A7CFA54FF8FA85A619E830ECFB0`

### `list-pool-ca`

List the current CA certificates in the CA pool.

Also display a 'hash' of CA serial, that you may use to compare others USS CA pool certificates list easily.

### `get-pool-ca`

Return all CA certificate in the current pool.

Can be used for debugging or to synchronize the set of CA certificates in a pool with others USS.

### `get-ca`

Return your own CA certificate .

Can be used for debugging or to synchronize the set of CA certificates in a pool with others USS.

### `destroy`

Destroy a certificate set. Be careful, there are no way to undo the command.
