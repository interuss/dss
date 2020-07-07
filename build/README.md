# Deploying a DSS instance

## Deployment options

This document describes how to deploy a production DSS instance to interoperate with other DSS instances in a DSS Region.

To run a local DSS instance for testing, evaluation, or development, see [dev/standalone_instance.md](dev/standalone_instance.md).

## Glossary

- DSS Region - an entire synchronized DSS, typically operated by multiple
  organizations.
- DSS instance - a single logical replica in a DSS Region.

## Preface

This doc provides a well-lit path for deploying the DSS and its dependencies (
namely CockroachDB) on Kubernetes. This is not a requirement, and a DSS instance
can join a cluster as long as it meets the requirements below.

## CockroachDB requirements

- Every CockroachDB node must advertise a unique and routable address.
  - It is preferred to use domain names with unique prefixes and homogenous
    suffixes, ie: 0.c.dss.interussplatform.com.
  - This allows wildcard usage in the CRDB certificates.
- Every DSS instance should run a minimum of 3 CockroachDB nodes, which
  ensures enough nodes are always available to support failovers and gradual
  rollouts.
- At least 3 CockroacbDB addresses must be shared with all participants.
  - If not using the recommended hostname prefix above, every CockroachDB
    hostname must be shared with every participant.
- Every DSS instance must supply and share their CockroachDB public
  certificate.
- All CockroachDB nodes must be run in secure mode, by supplying the
  `--certs-dir` and `--ca-key` flags.
  - Do not specify `--insecure`
- The ordering of the `--locality` flag keys must be the same across all
  CockroachDB nodes in the cluster.
- All sharing must currently happen out of band.

Note: we are investigating the use of service mesh frameworks to alleviate some
of this operational overhead.

## Prerequisites

Download & install the following tools to your workstation:

- Install tanka: https://tanka.dev/install. Until
  https://github.com/grafana/tanka/pull/162 makes it into a release, run:
  `go get -u github.com/grafana/tanka/cmd/tk@master`
- Install kubectl
  - Required to interact with kubernetes.
- docker
  - Required if building new images of the DSS.
- Cockroachdb
  - Required to generate new CockroachDB certificates.
- Google Cloud SDK (if deploying on GCP)
  - Required if deploying to Google Cloud.
- Golang
  - Required if developing the DSS codebase.
- Optional - [Jsonnet](https://github.com/google/jsonnet)
  - Recommended if editing the jsonnet templates.

## Docker images

The grpc-backend and http-gateway are the 2 main binaries for processing DSS
requests. These are both built and pushed to a docker registry of your choice.
You can easily find out how to push to a docker registry through a quick search.
All major cloud providers have a docker registry service, or you can set up your
own.

## Deploying the DSS on Kubernetes

Note: All DSS instances in the same cluster must point their ntpd at the
same NTP Servers. [CockroachDB recommends](https://www.cockroachlabs.com/docs/stable/recommended-production-settings.html#considerations)
using [Google's Public NTP](https://developers.google.com/time/) when
running in a multi-cloud environment.

This secion discusses deploying a Kubernetes service, although you can deploy
a DSS instance however you like as long as it meets the CockroachDB requirements
above. You can do this on any supported
[cloud provider](https://kubernetes.io/docs/concepts/cluster-administration/cloud-providers/)
or even on your own infrastructure. Consult the Kubernetes documentation for
which ever provider you choose.

1.  Create a new Kubernetes cluster as mentioned above. We recommend a new
    cluster for each DSS instance.

1.  Create a namespace for your Kubernetes cluster in which the DSS services will reside. This can be achieved with `kubectl create ns <NAMESPACE>`. Alternatively, you may want deploy in a namespace that already exists, so keep track of the namespace for later steps. The namespace name does not matter, though it will affect the DNS addresses associated with the certificates generated. One convention might be to reference a namespace named `production` to differentiate another cluster namespace called `development`. Another might be to name it `dss` to differentiate these services from the other namespaces within the cluster.

1.  Create your static IP addresses. How you do this depends on your cloud
    provider:

    - 1 for the HTTPS Gateway's Ingress. Note: if you're using Google Cloud
      the HTTPS Gateway Ingress needs to be created as a "Global" IP address.
    - If you want to be able to join other clusters, you'll need static IP's
      for each of your CockroachDB nodes (min of 3), and you'll need to create
      DNS entries with a common hostname suffix, ie: `*.db.interuss.com`

1.  Use the `make-certs.py` script in this directory to create certificates for
    the new CockroachDB cluster:

        ./make-certs.py --cluster <CLUSTER_CONTEXT> --namespace <NAMESPACE>
            [--node-address <ADDRESS> <ADDRESS> <ADDRESS> ...]
            [--ca-cert-to-join <CA_CERT_FILE>]

    1.  If you want other clusters to be able to connect to your cluster
        (including if you are joining an existing cluster) then `--node-address`
        needs to include all the hostnames and/or IP addresses that other
        CockroachDB clusters will use to connect to your cluster. Wildcard
        notation is supported, so you can use `*.<subdomain>.<domain>.com>`.
        The entries should be separated by spaces. These entries should
        correspond to the entries in main.jsonnet cockroach.nodeIPs.
    1.  If you are joining existing clusters you need their CA public cert,
        which is concatenated with yours. Set `--ca-cert-to-join` to a `ca.crt`
        file. Reach out to existing operators to request their public cert and
        node hostnames.
    1.  Note: If you are creating multiple clusters at once, and joining them
        together you likely want to copy the nth cluster's `ca.crt` into the the
        rest of the clusters, such that ca.crt is the same across all 3 clusters

1.  Build the docker file as described in the section below.

1.  Copy `cp -r deploy/examples/minimum` to
    `workspace/<CLUSTER_CONTEXT>` and fill in with your fields.

1.  If you are only turning up a single cluster for development, set
    metadata.single_cluster to true in your main.jsonnet.

1.  You can find the APIServer for spec.json by running:

    `kubectl config view -o jsonpath="{.clusters[?(@.name==\"<CLUSTER_CONTEXT>\")].cluster.server}"`

1.  Use the `apply-certs.sh` script in this directory to create secrets on the
    Kubernetes cluster containing the certificates and keys generated in the
    previous step.

        ./apply-certs.sh <CLUSTER_CONTEXT> <NAMESPACE>

1.  Follow the steps below before running the final command if you are joining
    existing clusters.

1. To start the profiling service for the grpc-backend and/or http-gateway then enter the service
   name for the profiler in your `main.jsonnet`. The name should be based on your cloud provider acceptable values. For Google cloud the name should meet the regex `^[a-z]([-a-z0-9_.]{0,253}[a-z0-9])?$`

1.  Run `tk apply workspace/<CLUSTER_CONTEXT>` to apply it to the
    cluster.

### Building Docker images

Set the environment variable `DOCKER_URL` to your docker registry url endpoint.

Use the `build.sh` script in this directory to build and push an image tagged
with the current date and git commit hash.

## Joining an existing CockroachDB cluster

Follow the steps above for creating a new CockroachDB cluster, but with the
following differences:

1.  In main.jsonnet, make sure you don't set shouldInit to true. This can
    initialize the data directories on you cluster, and prevent you from joining
    an existing cluster.

1.  In main.jsonnet, add the host:ports of existing CockroachDB nodes to the
    JoinExisting array. You should supply a minimum of 3 seed nodes to every
    CockroachDB node. These 3 nodes should be the same for every node (ie: every
    node points to node 0, 1, and 2). For external clusters you should point to
    a minimum of 3, or you can use a loadbalanced hostname or IP address
    of other clusters. You should do this for every cluster, including newly
    joined clusters. See CockroachDB's note on the
    [join flag](https://www.cockroachlabs.com/docs/stable/start-a-node.html#flags).

1.  You must run ./make-certs.py with the `--ca-cert-to-join` flag as described
    above to use the existing cluster's CA to sign your certificates.

1.  You must share ca.crt with the cluster(s) you are trying to join, and have
    them apply the new ca.crt, which now contains both your cluster and the
    original clusters public certs, to enable secure bi-directional
    communication.

1.  All of the original clusters must then perform a rolling restart of their
    cockroachdb pods to pick up the new certificates.
    `kubectl rollout restart statefulset/cockroachdb --namespace <NAMESPACE>`

## Enabling Istio

Istio provides better observability by using a sidecar proxy on every binary
that exports some default metrics, as well as enabling Istio tracing. Istio
also provides mTLS between all binaries. Enabling Istio is completely optional.
To enable Istio, simply change the `enable_istio` field in your metadata tuple
to `true`, then run `tk apply ...` as you would normally. 


## Enabling Prometheus Federation (Multi Cluster Monitoring)

The DSS uses [Prometheus](https://prometheus.io/docs/introduction/overview/) to gather metrics from the binaries deployed with this project, by scraping formatted metrics from an application's endpoint. [Prometheus Federation](https://prometheus.io/docs/prometheus/latest/federation/) enables you to easily monitor multiple clusters of the DSS that you operate, unifying all the metrics into a single Prometheus instance where you can build Grafana Dashboards for. Enabling Prometheus Federation is optional. To enable you need to do 2 things 
1. Externally expose the Prometheus service of the DSS clusters.
2. Deploy a "Global Prometheus" instance to unify metrics.

### Externally Exposing Prometheus
You will need to change the values in the `prometheus` fields in your metadata tuples:
1. `expose_external` set to `true`
2. [Optional] Supply a static external IP Address to `IP`
3. [Highly Recommended] Supply whitelists of [IP Blocks in CIDR form](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing), leaving an empty list mean everyone can publicly access your metrics.
4. Then Run `tk apply ...` to deploy the changes on your DSS clusters.

### Deploy "Global Prometheus" instance
1. Follow guide to deploy Prometheus https://prometheus.io/docs/introduction/first_steps/
2. The scrape rules for this global instance will scrape other prometheus `/federate` endpoint and rather simple, please look at the [example configuration](https://prometheus.io/docs/prometheus/latest/federation/#configuring-federation).

## Using the CockroachDB web UI

The CockroachDB web UI is not exposed publicly, but you can forward a port to
your local machine using the kubectl command:

### Create a user account

Pick a username and create an account:

    kubectl -n <NAMESPACE> exec cockroachdb-0 -ti -- \
        ./cockroach --certs-dir ./cockroach-certs \
        user set $USERNAME --password

### Access the web UI

    kubectl -n <NAMESPACE> port-forward cockroachdb-0 8080

Then go to https://localhost:8080. You'll have to ignore the HTTPS certificate
warning.
