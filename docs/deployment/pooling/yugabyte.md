# Pooling configuration with YugabyteDB

## Certificates management

### Introduction

The `dss-certs.py` helps you manage the set of certificates used for your DSS deployment.

Should this DSS beeing part of a pool, the script also provide some helpers to manage the set of CA certificates in the pool.

To run the script, just run `./dss-certs.py`. The python script don't require any dependencies, just a recent version of python 3.

### Quick start guide

#### Single DSS instance in minikube`

* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default init`
* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default apply`

#### Pool of 3 DSS instances in minikube, in namespace `default`, `ns2` and `ns3`

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

!!! Roll out restart required

### Operations

#### Common parameters

##### `--name`

The name of your DSS instance, that should identify it in a unique way. Used as main identifier for the set of certificates and in certificates.

Example: `dss-west-1`

##### `--organization`

The name of the organization managing the DSS Instance. Used in certificates generation. The combination of (name, organization) shall be unique in a cluster.

Example: `Interuss`

##### `--cluster-context`

The kubernetes context the script should use.

Example: `dss-local-cluster`

##### `--namespace`

The kubernetes namespace to use.

Example: `default`

##### `--nodes-count`

The number of yugabyte nodes of your DSS instance. Default to `3`.

#### `init`

Initializes the certificates for a new DSS instance including a CA, a client certificate and a certificate for each yugabyte node.

#### `apply`

Apply the current set of certificates to the kubernetes cluster. Shall be ran after each modification of the certificates, like addition / removal of CA in the pool, new `nodes-count` parameter.

#### `regenerate-nodes`

Generate missing nodes certificates. Useful if you want to add new nodes in your DSS Instance. Don't forget to set the `nodes-count` parameters.

#### `add-pool-ca`

Add a CA certificate(s) of another(s) DSS Instance to the set of trusted certificates.
Existing certificates are not added again.

You can set the file with certificate(s) with `--ca-file` or use stdin.

Don't forget to use the `apply` command to update certificate on your kubernetes cluster.

Examples:

* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default add-pool-ca < /tmp/new-dss-ca`
* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default --ca-file /tmp/new-dss-ca add-pool-ca`
* `./dss-certs.py --name test --cluster-context dss-local-cluster --namespace default get-pool-ca | ./dss-certs.py --name test2 --cluster-context dss-local-cluster --namespace namespace2 add-pool-ca`

#### `remove-pool-ca`

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

#### `list-pool-ca`

List the set of accepted CA certificates.

Also display a 'hash' of CA serial, that you may use to compare other DSS Instances list of CA certificates easily.

#### `get-pool-ca`

Return all CA certificate in the current pool.

Can be used for debugging or to synchronize the set of CA certificates in a pool with others USS.

#### `get-ca`

Return your own CA certificate .

Display the compiled CA certificate. Can be used for debugging or to synchronize the set of CA certificates in a pool with others USS.

#### `destroy`

Destroy a certificate set. Be careful, there are no way to undo the command.

## Deployment into a YugabyteDB pool

### First instance

Use [`dss-certs.py` script](#certificates-management) to create certificates for the Yugabyte nodes in this DSS instance.

Each DSS instance must set `yugabyte_external_nodes` with the list of each
others DSS instance Yugabyte master nodes public endpoints, and CA certificates
must be exchanged.

It's possible to have one DSS instance as starting point. In that case,
`yugabyte_external_nodes` will be empty and no CA exchange is needed.

!!! info
    Quick reminder for CA management:

    Each DSS instance should use `./dss-certs.sh init` To get the CA that should
    be sent to others instances, use `./dss-certs.sh get-ca` To import the CA of
    others DSS instance, use `./dss-certs.sh add-pool-ca` Finally, apply
    certificates on the kubernetes cluster with `./dss-certs.sh apply`

Ensure placement info is how you want it. See the section below for placement
requirements.

### Joining an existing pool with a new instance

They
will be joining an existing cluster, and they will need to request all CAs that
the pool is currently using (any one member of the pool may provide it). The
joining USS will also need a list of Yugabyte node addresses.

The joining USS must create his own CA with `./dss-certs.sh init` and retrieve
it with `./dss-certs.sh get-ca`. This certificate must be provided to each
existing DSS instance in the pool that will import it with `./dss-certs.sh
add-pool-ca` and `./dss-certs.sh apply`.

One of existing DSS instance shall provide to the joining USS all existing
certificate, using `./dss-certs.sh get-pool-ca`. The joining USS can import them
with `./dss-certs.sh add-pool-ca` and finally apply certificates with
`./dss-certs.sh apply`. As an alternative, each DSS instance can provide its
individual CA.

Participants shall ensure they work with a coherent set of certificate by
comparing the pool CA hash. It is displayed after adding certificates or using
the `./dss-certs.sh list-pool-ca`.

When CAs have been exchanged and configured everywhere, the joining participant
can bring his system online (e.g. by applying helm charts onto his cluster). The
`yugabyte_external_nodes` setting shall be set **before** starting the Yugabyte
cluster.

New nodes shall be allowed into the cluster. For each new Yugabyte master node,
the following command shall be run on one master node of one existing DSS
instance :

!!! warning
    The `master_addresses` in all commands below must include the Yugabyte master
    leader. Either always run commands in the cluster with the leader, or list all
    public addresses.

1. Connection to a master node:

    `kubectl exec -it yb-master-0 -- sh`

1. Addition of a new master node

    ``yb-admin -certs_dir_name /opt/certs/yugabyte/ -client_node_name=`hostname -f` -master_addresses yb-master-0.yb-masters.default.svc.cluster.local:7100,yb-master-1.yb-masters.default.svc.cluster.local:7100,yb-master-2.yb-masters.default.svc.cluster.local:7100 change_master_config ADD_SERVER [PUBLIC HOSTNAME] 7100``

The last command can be repeated as needed, however a small delay is needed for
the cluster to settle when adding a new node. If you get `Leader is not ready
for Config Change, can try again`, just try again.

You should have all masters listed in the web ui or using the
``yb-admin -certs_dir_name /opt/certs/yugabyte/ -client_node_name=`hostname -f` -master_addresses yb-master-0.yb-masters.default.svc.cluster.local:7100,yb-master-1.yb-masters.default.svc.cluster.local:7100,yb-master-2.yb-masters.default.svc.cluster.local:7100 list_all_masters``
command.

The tserver nodes will join automatically, using the list of provided master
nodes. They can be listed for confirmation in the web ui or using the
``yb-admin -certs_dir_name /opt/certs/yugabyte/ -client_node_name=`hostname -f` -master_addresses yb-master-0.yb-masters.default.svc.cluster.local:7100,yb-master-1.yb-masters.default.svc.cluster.local:7100,yb-master-2.yb-masters.default.svc.cluster.local:7100 list_all_tablet_servers``
command.

The pool should then be re-verified for functionality
by running the prober test on each DSS instance, and the
[interoperability test scenario](https://github.com/interuss/monitoring/blob/main/monitoring/uss_qualifier/scenarios/astm/netrid/v19/dss_interoperability.md)
on the full pool (including the newly-added instance).

Finally, the joining USS should provide its Yugabyte node addresses to all other
participants in the pool, and each other participant should add those addresses
to the `yugabyte_external_nodes` list their Yugabyte nodes will attempt to
contact upon restart.

Ensure placement info is how you want it. See the section below for placement
requirements.
