# Deploying a DSS



## Glossary

*   DSS cluster - an entire synchronized DSS, typically operated by multiple
    organizations
*   DSS node - a single logical replica in a synchronized cluster.


## Preface


This doc provides a well-lit path for deploying the DSS and its dependencies (
namely CockroachDB) on Kubernetes. This is not a requirement, and a DSS node
can join a cluster as long as it meets the requirements below.


## CockroachDB requirements


*   Every CockroachDB node must advertise a unique and routable address.
    *   It is preferred to use domain names with unique prefixes and homogenous
        suffixes, ie: 0.c.dss.interussplatform.com.
    *   This allows wildcard usage in the CRDB certificates.
*   Every DSS node should run a minimum of 3 CockroachDB nodes, which ensures
    enough nodes are always available to support failovers and gradual rollouts.
*   At least 3 CockroacbDB addresses must be shared with all participants.
    *   If not using the recommended hostname prefix above, every cockroachDB
        hostname must be shared with every participant.
*   Every DSS node must supply and share their CockroachDB public certificate
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

*   The [helm client](https://helm.sh/docs/using_helm/#installing-the-helm-client)
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


## Deploying the DSS on Kubernetes

This secion discusses deploying a Kubernetes service, although you can deploy 
a DSS node however you like as long as it meets the CockroachDB requirements
above. You can do this on any supported
[cloud provider](https://kubernetes.io/docs/concepts/cluster-administration/cloud-providers/)
or even on your own infrastructure.  Consult the Kubernetes documentation for
which ever provider you choose.

1.  Create a new Kubernetes cluster as mentioned above. We recommend a new
    cluster for each DSS node.

1.  Create your static IP addresses: 1 for each of your CockroachDB nodes (min
    of 3) and 1 for the HTTPS Gateway's Ingress.  How you do this depends on
    your cloud provider. Note: if you're using Google Cloud the HTTPS Gateway
    Ingress needs to be created as a "Global" IP address.

1.  Copy `values.yaml.template` to `values.yaml` and fill in the required fields
    at the top.

1.  Use the `make-certs.py` script in this directory to create certificates for
    the new CockroachDB cluster:

        ./make-certs.py [--node-address <ADDRESS> <ADDRESS> <ADDRESS> ...]
            [--ca-cert-to-join <CA_CERT_FILE>]

*   Make sure you supply the following flags if joining an existing cluster:
    *   --node-address needs to include all the hostnames and/or IP addresses 
        that other CockroachDB clusters will use to connect to your cluster.  It
        should include the addresses of your nodes as well.  Wildcard notation
        is supported, so you can use `*.<subdomain>.<domain>.com>`.  The
        entries should be separated by spaces.
    *   If you are joining existing clusters you need their CA public cert,
        which is concatenated with yours.  Set `--ca-cert-to-join` to a `ca.crt`
        file. Reach out to existing operators to request their public cert and
        node hostnames.

1.  Use the `apply-certs.sh` script in this directory to create secrets on the
    Kubernetes cluster containing the certificates and keys generated in the
    previous step.

        ./apply-certs.sh

1.  Run `helm template . > cockroachdb.yaml` to render the YAML.
1.  Run `kubectl apply -f cockroachdb.yaml` to apply it to the cluster.


## Joining an existing CockroachDB cluster


Follow the steps above for creating a new CockroachDB cluster, but with the
following differences:

1.  In values.yaml, be sure to set ClusterInit to false. This can initialize
    the data directories on you cluster, and prevent you from joining an
    existing cluster.
1.  In values.yaml, add the host:ports of existing CockroachDB nodes to the
    JoinExisting array.  You can use the loadbalanced hostnames or IP addresses
    of other clusters (the DBBalanced hostname/IP), or you can specify each node
    individually.
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

