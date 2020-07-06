# Deploying a DSS instance

## Deployment options

This document describes how to deploy a production DSS instance to interoperate
with other DSS instances in a DSS Region.

To run a local DSS instance for testing, evaluation, or development, see
[dev/standalone_instance.md](dev/standalone_instance.md).

## Glossary

- DSS Region - an entire synchronized DSS, typically operated by multiple
  organizations.
- DSS instance - a single logical replica in a DSS Region.

## Preface

This doc provides a well-lit path for deploying the DSS and its dependencies (
namely CockroachDB) on Kubernetes. The use of Kubernetes is not a requirement,
and a DSS instance can join a cluster as long as it meets the
[CockroachDB requirements below](#cockroachdb-requirements).

## Prerequisites

Download & install the following tools to your workstation:

- [Install kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) to
  interact with kubernetes
  - Confirm sucessful installation with `kubectl version --client` (should
    succeed from any working directory).
  - Note that kubectl can alternatively be installed via the Google Cloud SDK
   `gcloud` shell if using Google Cloud.
- [Install tanka](https://tanka.dev/install)
  - On Linux, after downloading the binary per instructions, run
    `sudo chmod +x /usr/local/bin/tk`
  - Confirm successful installation with `tk --version`
- If building new images of the DSS,
  [install Docker](https://docs.docker.com/get-docker/).
  - Confirm successful installation with `docker --version`
- If generating new CockroachDB certificates,
  [install CockroachDB](https://www.cockroachlabs.com/get-cockroachdb/)
  - These instructions assume CockroachDB Core.
  - You may need to run `sudo chmod +x /usr/local/bin/cockroach` after
    completing the installation instructions.
  - Confirm successful installation with `cockroach version`
- If deploying on Google Cloud,
  [install Google Cloud SDK](https://cloud.google.com/sdk/install)
  - Confirm successful installation with `gcloud version`
  - Run `gcloud init` to set up a connection to your account.
- If developing the DSS codebase,
  [install Golang](https://golang.org/doc/install)
  - Confirm successful installation with `go version`
- Optionally install [Jsonnet](https://github.com/google/jsonnet) if editing
  the jsonnet templates.

## Docker images

The grpc-backend and http-gateway are the two main binaries for processing DSS
requests. These are both built and pushed to a docker registry of your choice.
You can easily find out how to push to a docker registry through a quick search.
All major cloud providers have a docker registry service, or you can set up your
own.

To build these images (and, optionally, push them to a docker registry):

1. Set the environment variable `DOCKER_URL` to your docker registry url
endpoint.

    -   For Google Cloud, `DOCKER_URL` should be set similarly to as described
        [here](https://cloud.google.com/container-registry/docs/pushing-and-pulling#tag_the_local_image_with_the_registry_name),
        like `gcr.io/your-project-id` (do not include the image name;
        it will be appended by the build script)

1. Ensure you are logged into your docker registry service.

    -   For Google Cloud,
        [these](https://cloud.google.com/container-registry/docs/advanced-authentication#gcloud-helper)
        are the recommended instructions (`gcloud auth config-docker`). Ensure
        that
        [appropriate permissions are enabled](https://cloud.google.com/container-registry/docs/access-control)

1. Use the [`build.sh` script](./build.sh) in this directory to build and push
an image tagged with the current date and git commit hash.

    -   If using docker requires `sudo`, ensure `DOCKER_URL` is passed correctly
        with `sudo -E ./build.sh`.

## Deploying the DSS on Kubernetes

Note: All DSS instances in the same cluster must point their ntpd at the
same NTP Servers. [CockroachDB recommends](https://www.cockroachlabs.com/docs/stable/recommended-production-settings.html#considerations)
using [Google's Public NTP](https://developers.google.com/time/) when
running in a multi-cloud environment.

This section discusses deploying a Kubernetes service, although you can deploy
a DSS instance however you like as long as it meets the CockroachDB requirements
above. You can do this on any supported
[cloud provider](https://kubernetes.io/docs/concepts/cluster-administration/cloud-providers/)
or even on your own infrastructure. Consult the Kubernetes documentation for
your chosen provider.

1.  Create a new Kubernetes cluster. We recommend a new cluster for each DSS
    instance.  A reasonable cluster name might be `dss-us-prod`,
    `dss-ca-staging`, `dss-mx-integration`, etc.  The name of this cluster
    will be referred to as CLUSTER_CONTEXT for the remainder of this document.
    
    -  On Google Cloud, the recommended procedure to create a cluster is:
       -  In Google Cloud Platform, go to the Kubernetes Engine page and under
          Clusters click Create cluster.
       -  Name the cluster appropriately; e.g., `dss-us-prod`
       -  Select Zonal and [a compute-zone appropriate to your
          geography](https://cloud.google.com/compute/docs/regions-zones#available)
       -  For the "default-pool" node pool, check "Enable autoscaling" and enter
          a range of 2-10 nodes.
       -  In the "Nodes" bullet under "default-pool", select N2 series and
          n2-standard-8 for machine type.
       -  In the "Networking" bullet under "Clusters", ensure "Enable [VPC
          -native traffic](https://cloud.google.com/kubernetes-engine/docs/how-to/alias-ips)"
          is checked.

1.  Make sure correct cluster context is selected by printing the context
    name to the console: `kubectl config current-context`
    
    -  Record this value and use it for `CLUSTER_CONTEXT` below.

1.  Ensure the desired namespace is selected; the recommended
    namespace is simply `default` with one cluster per DSS instance.  Print the
    the current namespaces with `kubectl get namespace`.  Use the current
    namespace as the value for `NAMESPACE` below.
    
1.  Create static IP addresses: one for the HTTPS Gateway's ingress, and one
    for each CockroachDB node (minimum of 3) if you want to be able to join
    other clusters.

    -  If using Google Cloud, the HTTPS Gateway ingress needs to be created as
       a "Global" IP address.  IPv4 is recommended as IPv6 has not yet been
       tested.  Follow
       [these instructions](https://cloud.google.com/compute/docs/ip-addresses/reserve-static-external-ip-address#reserve_new_static)
       to reserve the static IP addresses.  Specifically (replacing
       CLUSTER_NAME as appropriate since static IP addresses are defined at
       the project level rather than the cluster level), e.g.:
       
         -  `gcloud compute addresses create CLUSTER_NAME-gateway --global --ip
         -version IPV4`
         -  `gcloud compute addresses create CLUSTER_NAME-crdb-0 --global --ip
         -version IPV4`
         -  `gcloud compute addresses create CLUSTER_NAME-crdb-1 --global --ip
         -version IPV4`
         -  `gcloud compute addresses create CLUSTER_NAME-crdb-2 --global --ip
         -version IPV4`

1.  Link static IP addresses to DNS entries.

    -  If joining other clusters, your CockroachDB nodes should have a common
       hostname suffix; e.g., `*.db.interuss.com`
    
    -  If using Google Cloud, see
       [these instructions](https://cloud.google.com/dns/docs/quickstart#create_a_new_record)
       to create DNS entries for the static IP addresses created above.  To list
       the IP addresses, use `gcloud compute addresses list`.

1.  Use [`make-certs.py` script](./make-certs.py) to create certificates for
    the new CockroachDB cluster:

        ./make-certs.py --cluster <CLUSTER_CONTEXT> --namespace <NAMESPACE>
            [--node-address <ADDRESS> <ADDRESS> <ADDRESS> ...]
            [--ca-cert-to-join <CA_CERT_FILE>]

    1.  `CLUSTER_CONTEXT` is the name of the cluster (see step 2 above).
    
    1.  `Each ADDRESS` is the DNS entry for a CockroachDB node.  To enable
        other clusters to connect to your cluster (including if joining an
        existing cluster) then `--node-address` must include all the hostnames
        hostnames and/or IP addresses that other CockroachDB clusters will use
        to connect to your cluster. Wildcard notation is supported, so you can
        use `*.<subdomain>.<domain>.com>`.  The entries should be separated by
        spaces.

    1.  If you are joining existing cluster(s) you need their CA public cert,
        which is concatenated with yours. Set `--ca-cert-to-join` to a `ca.crt`
        file.  Reach out to existing operators to request their public cert and
        node hostnames.  If not joining an existing cluster, omit this argument.

    1.  Note: If you are creating multiple clusters at once, and joining them
        together you likely want to copy the nth cluster's `ca.crt` into the the
        rest of the clusters, such that ca.crt is the same across all clusters.

1.  If joining an existing cluster, share ca.crt with the cluster(s) you are
    trying to join, and have them apply the new ca.crt, which now contains both
    your cluster and the original clusters public certs, to enable secure bi
    -directional communication.
    
    - All of the original clusters must perform a rolling restart of their
      cockroachdb pods to pick up the new certificates:
      
      `kubectl rollout restart statefulset/cockroachdb --namespace <NAMESPACE>`

1.  Ensure the Docker images are built according to the instructions in the
    [previous section](#docker-images).

1.  From this working directory,
    `cp -r deploy/examples/minimum/* workspace/<CLUSTER_CONTEXT>`.  Note that
    the `workspace/<CLUSTER_CONTEXT>` folder should have already been created
    by the `make_certs.py` script.

1. Edit `workspace/<CLUSTER_CONTEXT>/main.jsonnet` and replace all `VAR_*`
   instances with appropriate values:
   
   -   VAR_NAMESPACE: Same <NAMESPACE> used in the make-certs.py (and
       apply-certs.sh) scripts.
   
   -   VAR_CLUSTER_CONTEXT: Same <CLUSTER_CONTEXT> name of the cluster used in
       the `make-certs.py` and `apply-certs.sh` scripts.
   
   -   VAR_CRDB_HOSTNAME_SUFFIX: The domain name suffix shared by all of your
       CockroachDB nodes.  For instance, if your CRDB nodes were addressable at
       `node1.db.example.com`, `node2.db.example.com`, and
       `node3.db.example.com`, then VAR_CRDB_HOSTNAME_SUFFIX would be
       `db.example.com`.

   -   VAR_CRDB_LOCALITY: Unique name for your cluster.  TODO
   
   -   VAR_CRDB_NODE_IPn: **Numeric** IP address of nth CRDB node (add more
       entries if you have more than 3 CRDB nodes).  Example: `1.1.1.1`
       
   -   VAR_SHOULD_INIT: Set to `false` if joining an existing cluster, `true`
       otherwise.  When set `true`, this can initialize the data directories
       on your cluster, and prevent you from joining an existing cluster
       .  TODO: should this be set false immediately after initial application?
       
   -   VAR_EXISTING_CRDB_NODEn: Fully-qualified domain name of existing CRDB
       nodes if you are joining an existing cluster.  If more than three are
       available, add additional entries.  If not joining an existing cluster,
       remove this entire `JoinExisting:` line.
       
       - You should supply a minimum of 3 seed nodes to every CockroachDB node.
         These 3 nodes should be the same for every node (ie: every node points
         to node 0, 1, and 2). For external clusters you should point to a
         minimum of 3, or you can use a loadbalanced hostname or IP address of
         other clusters. You should do this for every cluster, including newly
         joined clusters. See CockroachDB's note on the
         [join flag](https://www.cockroachlabs.com/docs/stable/start-a-node.html#flags).
   
   -   VAR_INGRESS_NAME: If using Google Kubernetes Engine, set this to the
       TODO
   
   -   VAR_DOCKER_IMAGE_NAME: Full name of the docker image built in the
       section above.  `build.sh` prints this name as the last thing it does
       when run with `DOCKER_URL` set.  It should look something like
       `gcr.io/your-project-id/dss:2020-07-01-46cae72cf`.
       
   -   VAR_APP_HOSTNAME: Fully-qualified domain name of your HTTPS Gateway
       ingress endpoint.  For example, `dss.example.com`.
       
   -   Note that VAR_DOCKER_IMAGE_NAME is used in two places.
   
   -   If you are only turning up a single cluster for development, you
       may optionally change `single_cluster` to `true`.

1.  Edit workspace/<CLUSTER_CONTEXT>/spec.json and replace all VAR_*
    instances with appropriate values:
   
    -   VAR_API_SERVER: Determine this value with the command:
    
        `kubectl config view -o jsonpath="{.clusters[?(@.name==\"<CLUSTER_CONTEXT>\")].cluster.server}"`
        
        - Note that `<CLUSTER_CONTEXT>` should be replaced with your actual
         `CLUSTER_CONTEXT` value prior to executing the above command.
   
    -   VAR_NAMESPACE: See previous section.

1.  Use the [`apply-certs.sh` script](apply-certs.sh) to create secrets on the
    Kubernetes cluster containing the certificates and keys generated in the
    previous step.

        ./apply-certs.sh <CLUSTER_CONTEXT> <NAMESPACE>

1.  To start the profiling service for the grpc-backend and/or http-gateway
    then enter the service name for the profiler at `metadata_base.libsonnet`.
    The name should be based on your cloud provider acceptable values. For
    Google cloud the name should meet the regex
    `^[a-z]([-a-z0-9_.]{0,253}[a-z0-9])?$`

1.  Run `tk apply workspace/<CLUSTER_CONTEXT>` to apply it to the
    cluster.

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

## CockroachDB requirements

- Every CockroachDB node must advertise a unique and routable address.
  - The use of domain names with unique prefixes and homogenous suffixes, e.g.:
    0.c.dss.interussplatform.com, is preferred as this allows wildcard usage in
    the CRDB certificates.
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

## Enabling Istio

Istio provides better observability by using a sidecar proxy on every binary
that exports some default metrics, as well as enabling Istio tracing. Istio
also provides mTLS between all binaries. Enabling Istio is completely optional.
To enable Istio, simply change the `enable_istio` field in your metadata tuple
to `true`, then run `tk apply ...` as you would normally. 


## Enabling Prometheus Federation (Multi Cluster Monitoring)

The DSS uses [Prometheus](https://prometheus.io/docs/introduction/overview/) to
gather metrics from the binaries deployed with this project, by scraping
formatted metrics from an application's endpoint.
[Prometheus Federation](https://prometheus.io/docs/prometheus/latest/federation/)
enables you to easily monitor multiple clusters of the DSS that you operate,
unifying all the metrics into a single Prometheus instance where you can build
Grafana Dashboards for. Enabling Prometheus Federation is optional. To enable
you need to do 2 things:
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
your local machine using kubectl:

### Create a user account

Pick a username and create an account:

    kubectl -n <NAMESPACE> exec cockroachdb-0 -ti -- \
        ./cockroach --certs-dir ./cockroach-certs \
        user set $USERNAME --password

### Access the web UI

    kubectl -n <NAMESPACE> port-forward cockroachdb-0 8080

Then go to https://localhost:8080. You'll have to ignore the HTTPS certificate
warning.
