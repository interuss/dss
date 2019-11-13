# Deploying a DSS



## Glossary

*   DSS Region - an entire synchronized DSS, typically operated by multiple
    organizations
*   DSS instance - a single logical replica in a synchronized cluster.


## Preface


This doc provides a well-lit path for deploying the DSS and its dependencies (
namely CockroachDB) on Kubernetes. This is not a requirement, and a DSS instance
can join a cluster as long as it meets the requirements below.


## CockroachDB requirements


*   Every CockroachDB node must advertise a unique and routable address.
    *   It is preferred to use domain names with unique prefixes and homogenous
        suffixes, ie: 0.c.dss.interussplatform.com.
    *   This allows wildcard usage in the CRDB certificates.
*   Every DSS instance should run a minimum of 3 CockroachDB nodes, which
    ensures enough nodes are always available to support failovers and gradual
    rollouts.
*   At least 3 CockroacbDB addresses must be shared with all participants.
    *   If not using the recommended hostname prefix above, every CockroachDB
        hostname must be shared with every participant.
*   Every DSS instance must supply and share their CockroachDB public
    certificate.
*   All CockroachDB nodes must be run in secure mode, by supplying the
    `--certs-dir` and `--ca-key` flags.
    * Do not specify `--insecure`
*   The ordering of the `--locality` flag keys must be the same across all
    CockroachDB nodes in the cluster.
*   All sharing must currently happen out of band.

Note: we are investigating the use of service mesh frameworks to alleviate some
of this operational overhead.


## Prerequisites


Download & install the following tools to your workstation:

*   The [kubecfg client](https://github.com/bitnami/kubecfg#install)
    *   Required if deploying using the defined Kubernetes templates.
*   kubectl
    *   Required if deploying with Kubernetes.
*   docker
    *   Required if building new images of the DSS.
*   Cockroachdb
    *   Required to generate new CockroachDB certificates.
*   Google Cloud SDK (if deploying on GCP)
    * Required if deploying to Google Cloud.
*   Golang.
    *   Required if developing the DSS codebase.
*   Optional - [Jsonnet](https://github.com/google/jsonnet)
    * Recommended if editing the jsonnet templates.


## Docker images


The grpc-backend and http-gateway are the 2 main binaries for processing DSS 
requests. These are both built and pushed to a docker registry of your choice.
You can easily find out how to push to a docker registry through a quick search.
All major cloud providers have a docker registry service, or you can set up your
own.


### Building Docker images


Set the environment variable `DOCKER_URL` to your docker registry url endpoint.

Use the `build.sh` script in this directory to build and push an image tagged
with the current date and git commit hash.


## Running locally


Simply run the `run-locally.sh` script in this directory.


## Deploying the DSS on Kubernetes


Note: All DSS instances in the same cluster must point their ntpd at the
same NTP Servers. [CockroachDB recommends](https://www.cockroachlabs.com/docs/stable/recommended-production-settings.html#considerations)
using [Google's Public NTP](https://developers.google.com/time/) when
running in a multi-cloud environment.

This secion discusses deploying a Kubernetes service, although you can deploy 
a DSS instance however you like as long as it meets the CockroachDB requirements
above. You can do this on any supported
[cloud provider](https://kubernetes.io/docs/concepts/cluster-administration/cloud-providers/)
or even on your own infrastructure.  Consult the Kubernetes documentation for
which ever provider you choose.

1.  Create a new Kubernetes cluster as mentioned above. We recommend a new
    cluster for each DSS instance.

1.  Create your static IP addresses. How you do this depends on your cloud
    provider:
    *   1 for the HTTPS Gateway's Ingress. Note: if you're using Google Cloud
        the HTTPS Gateway Ingress needs to be created as a "Global" IP address.
    *   If you want to be able to join other clusters, you'll need static IP's 
        for each of your CockroachDB nodes (min of 3).
    *   [OPTIONAL] 1 for a loadbalancer across your CockroachDB nodes to provide
        a single join target for other users.

1.  Copy `deploy/examples/minimum.jsonnet` to `dss.jsonnet` and fill in with
    your fields.

1.  Use the `make-certs.py` script in this directory to create certificates for
    the new CockroachDB cluster:

        ./make-certs.py [--node-address <ADDRESS> <ADDRESS> <ADDRESS> ...]
            [--ca-cert-to-join <CA_CERT_FILE>]

    1.  If you want other clusters to be able to connect to your cluster
        (including if you are joining an existing cluster) then `--node-address`
        needs to include all the hostnames and/or IP addresses that other
        CockroachDB clusters will use to connect to your cluster. Wildcard
        notation is supported, so you can use `*.<subdomain>.<domain>.com>`.
        The entries should be separated by spaces. These entries should
        correspond to the entries in dss.jsonnet cockroach.nodeIPs.
    1.  If you are joining existing clusters you need their CA public cert,
        which is concatenated with yours.  Set `--ca-cert-to-join` to a `ca.crt`
        file. Reach out to existing operators to request their public cert and
        node hostnames.

1.  Use the `apply-certs.sh` script in this directory to create secrets on the
    Kubernetes cluster containing the certificates and keys generated in the
    previous step.

        ./apply-certs.sh

1.  Run `kubecfg show <your_leaf_file.jsonnet> > dss.yaml` to render the YAML.
1.  Run `kubectl apply -f cockroachdb.yaml` to apply it to the cluster.


## Joining an existing CockroachDB cluster


Follow the steps above for creating a new CockroachDB cluster, but with the
following differences:

1.  In dss.jsonnet, make sure you don't set shouldInit to true. This can
    initialize the data directories on you cluster, and prevent you from joining
    an existing cluster.
1.  In dss.jsonnet, add the host:ports of existing CockroachDB nodes to the
    JoinExisting array.  You should supply a minimum of 3 seed nodes to every 
    CockroachDB node. These 3 nodes should be the same for every node (ie: every
    node points to node 0, 1, and 2).  For external clusters you should point to
    a minimum of 3, or you can use the loadbalanced hostnames or IP addresses
    of other clusters (the DBBalanced hostname/IP). You should do this for every
    cluster, including newly joined clusters. See CockroachDB's note on the 
    [join flag](https://www.cockroachlabs.com/docs/stable/start-a-node.html#flags).
1.  You must run ./make-certs.py with the `--ca-cert-to-join` flag as described
    above to use the existing cluster's CA to sign your certificates.


## Using the CockroachDB web UI


The CockroachDB web UI is not exposed publicly, but you can forward a port to
your local machine using the kubectl command:


### Create a user account


Pick a username and create an account:

    kubectl -n dss-main exec cockroachdb-0 -ti -- \
        ./cockroach --certs-dir ./cockroach-certs \
        user set $USERNAME --password


### Access the web UI


    kubectl -n dss-main port-forward cockroachdb-0 8080

Then go to https://localhost:8080.  You'll have to ignore the HTTPS certificate
warning.

