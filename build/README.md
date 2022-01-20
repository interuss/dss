# Deploying a DSS instance

## Deployment options

This document describes how to deploy a production-style DSS instance to
interoperate with other DSS instances in a DSS pool.

To run a local DSS instance for testing, evaluation, or development, see
[dev/standalone_instance.md](dev/standalone_instance.md).

To create or join a pool consisting of multiple interoperable DSS instances, see
[information on pooling](pooling.md).

## Glossary

- DSS Region - A region in which a single, unified airspace representation is
  presented by one or more interoperable DSS instances, each instance typically
  operated by a separate organization.  A specific environment (for example,
  "production" or "staging") in a particular DSS Region is called a "pool".
- DSS instance - a single logical replica in a DSS pool.

## Preface

This doc describes a procedure for deploying the DSS and its dependencies
(namely CockroachDB) via Kubernetes. The use of Kubernetes is not a requirement,
and a DSS instance can join a cluster as long as it meets the
[CockroachDB requirements below](#cockroachdb-requirements).

## Prerequisites

Download & install the following tools to your workstation:

- If deploying on Google Cloud,
  [install Google Cloud SDK](https://cloud.google.com/sdk/install)
  - Confirm successful installation with `gcloud version`
  - Run `gcloud init` to set up a connection to your account.
  - `kubectl` can be installed from `gcloud` instead of via the method below.
- [Install kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) to
  interact with kubernetes
  - Confirm successful installation with `kubectl version --client` (should
    succeed from any working directory).
  - Note that kubectl can alternatively be installed via the Google Cloud SDK
   `gcloud` shell if using Google Cloud.
- [Install tanka](https://tanka.dev/install)
  - On Linux, after downloading the binary per instructions, run
    `sudo chmod +x /usr/local/bin/tk`
  - Confirm successful installation with `tk --version`
- [Install Docker](https://docs.docker.com/get-docker/).
  - Confirm successful installation with `docker --version`
- [Install CockroachDB](https://www.cockroachlabs.com/get-cockroachdb/) to
  generate CockroachDB certificates.
  - These instructions assume CockroachDB Core.
  - You may need to run `sudo chmod +x /usr/local/bin/cockroach` after
    completing the installation instructions.
  - Confirm successful installation with `cockroach version`
- If developing the DSS codebase,
  [install Golang](https://golang.org/doc/install)
  - Confirm successful installation with `go version`
- Optionally install [Jsonnet](https://github.com/google/jsonnet) if editing
  the jsonnet templates.

## Docker images

The application logic of the DSS is located in core-service and translation
between external HTTPS requests and internal gRPC requests to core-service is
accomplished with http-gateway.  Both of these binaries are provided in a single
Docker image which is built locally and then pushed to a Docker registry of your
choice.  All major cloud providers have a docker registry service, or you can
set up your own.

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
        are the recommended instructions (`gcloud auth configure-docker`).
        Ensure that
        [appropriate permissions are enabled](https://cloud.google.com/container-registry/docs/access-control).

1. Use the [`build.sh` script](./build.sh) in this directory to build and push
   an image tagged with the current date and git commit hash.

1. Note the two VAR_* values printed at the end of the script.

## Deploying a DSS instance via Kubernetes

This section discusses deploying a Kubernetes service, although you can deploy
a DSS instance however you like as long as it meets the CockroachDB requirements
above. You can do this on any supported
[cloud provider](https://kubernetes.io/docs/concepts/cluster-administration/cloud-providers/)
or even on your own infrastructure. Consult the Kubernetes documentation for
your chosen provider.

If you can augment this documentation with specifics for another cloud provider,
a PR to that effect would be greatly appreciated.

1.  Create a new Kubernetes cluster. We recommend a new cluster for each DSS
    instance.  A reasonable cluster name might be `dss-us-prod-e4a` (where `e4a`
    is a zone identifier abbreviation), `dss-ca-staging`,
    `dss-mx-integration-sae1a`, etc.  The name of this cluster will be combined
    with other information by Kubernetes to generate a longer cluster context
    ID.

    -  On Google Cloud, the recommended procedure to create a cluster is:
       -  In Google Cloud Platform, go to the Kubernetes Engine page and under
          Clusters click Create cluster.
       -  Name the cluster appropriately; e.g., `dss-us-prod`
       -  Select Zonal and [a compute-zone appropriate to your
          geography](https://cloud.google.com/compute/docs/regions-zones#available)
       -  For the "default-pool" node pool:
          - Enter 3 for number of nodes.
          -  In the "Nodes" bullet under "default-pool", select N2 series and
             n2-standard-4 for machine type.
       -  In the "Networking" bullet under "Clusters", ensure "Enable [VPC
          -native traffic](https://cloud.google.com/kubernetes-engine/docs/how-to/alias-ips)"
          is checked.

1.  Make sure correct cluster context is selected by printing the context
    name to the console: `kubectl config current-context`

    -  Record this value and use it for `$CLUSTER_CONTEXT` below; perhaps:
       `export CLUSTER_CONTEXT=$(kubectl config current-context)`

    - On Google Cloud, first configure kubectl to interact with the cluster
      created above with
      [these instructions](https://cloud.google.com/kubernetes-engine/docs/quickstart).
      Specifically:
       - `gcloud config set project your-project-id`
       - `gcloud config set compute/zone your-compute-zone`
       - `gcloud container clusters get-credentials your-cluster-name`

1.  Ensure the desired namespace is selected; the recommended
    namespace is simply `default` with one cluster per DSS instance.  Print the
    the current namespaces with `kubectl get namespace`.  Use the current
    namespace as the value for `$NAMESPACE` below; perhaps use an environment
    variable for convenience: `export NAMESPACE=<your namespace>`.

1.  Create static IP addresses: one for the HTTPS Gateway's ingress, and one
    for each CockroachDB node if you want to be able to interact with other
    DSS instances.

    -  If using Google Cloud, the HTTPS Gateway ingress needs to be created as
       a "Global" IP address, but the CRDB ingresses as "Regional" IP addresses.
       IPv4 is recommended as IPv6 has not yet been tested.  Follow
       [these instructions](https://cloud.google.com/compute/docs/ip-addresses/reserve-static-external-ip-address#reserve_new_static)
       to reserve the static IP addresses.  Specifically (replacing
       CLUSTER_NAME as appropriate since static IP addresses are defined at
       the project level rather than the cluster level), e.g.:

         -  `gcloud compute addresses create ${CLUSTER_NAME}-gateway --global --ip-version IPV4`
         -  `gcloud compute addresses create ${CLUSTER_NAME}-crdb-0 --region $REGION`
         -  `gcloud compute addresses create ${CLUSTER_NAME}-crdb-1 --region $REGION`
         -  `gcloud compute addresses create ${CLUSTER_NAME}-crdb-2 --region $REGION`

1.  Link static IP addresses to DNS entries.

    -  Your CockroachDB nodes should have a common hostname suffix; e.g.,
       `*.db.interuss.com`.  Recommended naming is
       `0.db.yourdeployment.yourdomain.com`,
       `1.db.yourdeployment.yourdomain.com`, etc.

    -  If using Google Cloud, see
       [these instructions](https://cloud.google.com/dns/docs/quickstart#create_a_new_record)
       to create DNS entries for the static IP addresses created above.  To list
       the IP addresses, use `gcloud compute addresses list`.

1.  Use [`make-certs.py` script](./make-certs.py) to create certificates for
    the new CockroachDB cluster:

        ./make-certs.py --cluster $CLUSTER_CONTEXT --namespace $NAMESPACE
            [--node-address <ADDRESS> <ADDRESS> <ADDRESS> ...]
            [--ca-cert-to-join <CA_CERT_FILE>]

    1.  `$CLUSTER_CONTEXT` is the name of the cluster (see step 2 above).

    1.  `Each ADDRESS` is the DNS entry for a CockroachDB node.  To enable
        other clusters to connect to your cluster (including if joining an
        existing cluster) then `--node-address` must include all the hostnames
        hostnames and/or IP addresses that other CockroachDB clusters will use
        to connect to your cluster. Wildcard notation is supported, so you can
        use `*.<subdomain>.<domain>.com>`.  The entries should be separated by
        spaces.

    1.  If you are joining existing cluster(s) you need their CA public cert,
        which will be concatenated with yours. Set `--ca-cert-to-join` to a
        `ca.crt` file.  Reach out to existing operators to request their public
        cert and node hostnames.  If not joining an existing cluster, omit this
        argument.

    1.  Note: If you are creating multiple clusters at once, and joining them
        together you likely want to copy the nth cluster's `ca.crt` into the
        rest of the clusters, such that ca.crt is the same across all clusters.

1.  If joining an existing cluster, share ca.crt with the cluster(s) you are
    trying to join, and have them apply the new ca.crt, which now contains both
    your cluster and the original clusters public certs, to enable secure bi
    -directional communication.  The original cluster, upon receipt of the
    combined ca.crt from the joining cluster, should perform the actions below.
    While they are performing those actions, you may continue with the
    instructions.

    1. Overwrite its existing ca.crt with the new ca.crt provided by the joining
       cluster.
    1. Upload the new ca.crt to its cluster using
       `./apply-certs.sh $CLUSTER_CONTEXT $NAMESPACE`
    1. Restart their CockroachDB pods to recognize the updated ca.crt:
       `kubectl rollout restart statefulset/cockroachdb --namespace $NAMESPACE`
    1. Inform you when their CockroachDB pods have finished restarting
       (typically around 10 minutes)

1.  Ensure the Docker images are built according to the instructions in the
    [previous section](#docker-images).

1.  From this working directory,
    `cp -r deploy/examples/minimum/* workspace/$CLUSTER_CONTEXT`.  Note that
    the `workspace/$CLUSTER_CONTEXT` folder should have already been created
    by the `make-certs.py` script.

1.  If providing a .pem file directly as the public key to validate incoming
    access tokens, copy it to [dss/build/jwt-public-certs](./jwt-public-certs).
    Public key specification by JWKS is preferred; if using the JWKS approach
    to specify the public key, skip this step.

1.  Edit `workspace/$CLUSTER_CONTEXT/main.jsonnet` and replace all `VAR_*`
    instances with appropriate values:

    1.  `VAR_NAMESPACE`: Same `$NAMESPACE` used in the make-certs.py (and
        apply-certs.sh) scripts.

    1.  `VAR_CLUSTER_CONTEXT`: Same $CLUSTER_CONTEXT used in the `make-certs.py`
        and `apply-certs.sh` scripts.

    1.  `VAR_ENABLE_SCD`: Set this boolean true to enable strategic conflict
        detection functionality (currently an R&D project tracking an initial
        draft of the upcoming ASTM standard).

    1.  `VAR_CRDB_HOSTNAME_SUFFIX`: The domain name suffix shared by all of your
        CockroachDB nodes.  For instance, if your CRDB nodes were addressable at
        `0.db.example.com`, `1.db.example.com`, and `2.db.example.com`, then
        VAR_CRDB_HOSTNAME_SUFFIX would be `db.example.com`.

    1.  `VAR_CRDB_LOCALITY`: Unique name for your cluster.  Currently, we
        recommend "<ORG_NAME>_<CLUSTER_NAME>", and the `=` character is not
        allowed.  However, any unique (among all other participating DSS
        instances) value is acceptable.

    1.  `VAR_CRDB_NODE_IPn`: IP address (**numeric**) of nth CRDB node (add more
        entries if you have more than 3 CRDB nodes).  Example: `1.1.1.1`

    1.  `VAR_SHOULD_INIT`: Set to `false` if joining an existing cluster, `true`
        if creating the first DSS instance for a pool.  When set `true`, this
        can initialize the data directories on your cluster, and prevent you
        from joining an existing cluster.

    1.  `VAR_EXTERNAL_CRDB_NODEn`: Fully-qualified domain name of existing CRDB
        nodes if you are joining an existing cluster.  If more than three are
        available, add additional entries.  If not joining an existing cluster,
        comment out this `JoinExisting:` line.

        - You should supply a minimum of 3 seed nodes to every CockroachDB node.
          These 3 nodes should be the same for every node (ie: every node points
          to node 0, 1, and 2). For external clusters you should point to a
          minimum of 3, or you can use a loadbalanced hostname or IP address of
          other clusters. You should do this for every cluster, including newly
          joined clusters. See CockroachDB's note on the
          [join flag](https://www.cockroachlabs.com/docs/stable/start-a-node.html#flags).

    1.  `VAR_INGRESS_NAME`: If using Google Kubernetes Engine, set this to the
        the name of the gateway static IP address created above (e.g.,
        `CLUSTER_NAME-gateway`).

    1.  `VAR_DOCKER_IMAGE_NAME`: Full name of the docker image built in the
        section above.  `build.sh` prints this name as the last thing it does
        when run with `DOCKER_URL` set.  It should look something like
        `gcr.io/your-project-id/dss:2020-07-01-46cae72cf`.

    1.  `VAR_APP_HOSTNAME`: Fully-qualified domain name of your HTTPS Gateway
        ingress endpoint.  For example, `dss.example.com`.

    1.  `VAR_PUBLIC_KEY_PEM_PATH`: If providing a .pem file directly as the
        public key to validate incoming access tokens, specify the name of this
        .pem file here as `/public-certs/YOUR-KEY-NAME.pem` replacing
        YOUR-KEY-NAME as appropriate.  For instance, if using the provided
        [`us-demo.pem`](./jwt-public-certs/us-demo.pem), use the path
        `/public-certs/us-demo.pem`.  Note that your .pem file must have been
        copied into [`jwt-public-certs`](./jwt-public-certs) in an earlier step.

        - If providing an access token public key via JWKS, provide a blank
          string for this parameter.

    1.  `VAR_JWKS_ENDPOINT`: If providing the access token public key via JWKS,
        specify the JWKS endpoint here.  Example:
        `https://auth.example.com/.well-known/jwks.json`

        - If providing a .pem file directly as the public key to valid incoming access tokens, provide a blank string for this parameter.

    1.  `VAR_JWKS_KEY_ID`: If providing the access token public key via JWKS,
        specify the `kid` (key ID) of they appropriate key in the JWKS file
        referenced above.

        - If providing a .pem file directly as the public key to valid incoming access tokens, provide a blank string for this parameter.

    1.  `VAR_SCHEMA_MANAGER_IMAGE_NAME`: Full name of the schema manager docker
        image built in the section above.  `build.sh` prints this name as the
        last thing it does when run with `DOCKER_URL` set.  It should look
        something like `gcr.io/your-project-id/db-manager:2020-07-01-46cae72cf`.

    -   Note that `VAR_DOCKER_IMAGE_NAME` is used in two places.

    -   If you are only turning up a single cluster for development, you
        may optionally change `single_cluster` to `true`.

1.  Edit workspace/$CLUSTER_CONTEXT/spec.json and replace all VAR_*
    instances with appropriate values:

    1.  VAR_API_SERVER: Determine this value with the command:

        `echo $(kubectl config view -o jsonpath="{.clusters[?(@.name==\"$CLUSTER_CONTEXT\")].cluster.server}")`

        - Note that `$CLUSTER_CONTEXT` should be replaced with your actual
         `CLUSTER_CONTEXT` value prior to executing the above command if you
         have not defined a `CLUSTER_CONTEXT` environment variable.

    1.  VAR_NAMESPACE: See previous section.

1.  Use the [`apply-certs.sh` script](apply-certs.sh) to create secrets on the
    Kubernetes cluster containing the certificates and keys generated in the
    previous step.

        ./apply-certs.sh $CLUSTER_CONTEXT $NAMESPACE

1.  Run `tk apply workspace/$CLUSTER_CONTEXT` to apply it to the
    cluster.

    - If you are joining an existing cluster, do not execute this command until
      the existing cluster confirms that their CockroachDB pods have finished
      their rolling restarts.

1.  Wait for services to initialize.  Verify that basic services are functioning
    by navigating to https://your-gateway-domain.com/healthy.

    - On Google Cloud, the highest-latency operation is provisioning of the
      HTTPS certificate which generally takes 10-45 minutes.  To track this
      progress:
      - Go to the "Services & Ingress" left-side tab from the Kubernetes Engine
        page.
      - Click on the `https-ingress` item (filter by just the cluster of
        interest if you have multiple clusters in your project).
      - Under the "Ingress" section for Details, click on the link corresponding
        with "Load balancer".
      - Under Frontend for Details, the Certificate column for HTTPS protocol
        will have an icon next to it which will change to a green checkmark when
        provisioning is complete.
      - Click on the certificate link to see provisioning progress.
      - If everything indicates OK and you still receive a cipher mismatch error
        message when attempting to visit /healthy, wait an additional 5 minutes
        before attempting to troubleshoot further.

1.  If joining an existing cluster, share your CRDB node addresses with the
    operator of the existing cluster.  They will add these node addresses to
    JoinExisting where `VAR_CRDB_EXTERNAL_NODEn` is indicated in the minimum
    example, and then perform another rolling restart of their CockroachDB pods:

    `kubectl rollout restart statefulset/cockroachdb --namespace $NAMESPACE`

## Pooling

See [the pooling documentation](pooling.md).

## Tools

### Grafana / Prometheus

By default, an instance of Grafana and Prometheus are deployed along with the
core DSS services; this combination allows you to view (Grafana) CRDB metrics
(collected by Prometheus).  To view Grafana, first ensure that the appropriate
cluster context is selected (`kubectl config current-context`).  Then, run the
following command:

```shell script
kubectl get pod | grep grafana | awk '{print $1}' | xargs -I {} kubectl port-forward {} 3000
```

While that command is running, open a browser and navigate to
[http://localhost:3000](http://localhost:3000).  The default username is `admin`
with a default password of `admin`.  Click the magnifying glass on the left side
to select a dashboard to view.

### Istio

Istio provides better observability by using a sidecar proxy on every binary
that exports some default metrics, as well as enabling Istio tracing. Istio
also provides mTLS between all binaries. Enabling Istio is completely optional.
To enable Istio, simply set the `enable_istio` field in your metadata tuple to
`true`, then run `tk apply ...` as you would normally.

### Prometheus Federation (Multi Cluster Monitoring)

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

#### Externally Exposing Prometheus
You will need to change the values in the `prometheus` fields in your metadata tuples:
1. `expose_external` set to `true`
2. [Optional] Supply a static external IP Address to `IP`
3. [Highly Recommended] Supply whitelists of [IP Blocks in CIDR form](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing), leaving an empty list mean everyone can publicly access your metrics.
4. Then Run `tk apply ...` to deploy the changes on your DSS clusters.

#### Deploy "Global Prometheus" instance
1. Follow guide to deploy Prometheus https://prometheus.io/docs/introduction/first_steps/
2. The scrape rules for this global instance will scrape other prometheus `/federate` endpoint and rather simple, please look at the [example configuration](https://prometheus.io/docs/prometheus/latest/federation/#configuring-federation).

## Troubleshooting

### Check if the CockroachDB service is exposed

Unless specified otherwise in a deployment configuration, CockroachDB
communicates on port 26257.  To check whether this port is open from Mac or
Linux, e.g.: `nc -zvw3 0.db.dss.your-region.your-domain.com 26257`.  Or, search
for a "port checker" web page/app.  Port 26257 will be open on a working
CockroachDB node.

A standard TLS diagnostic may also be run on this hostname:port combination and
all results should be valid except Trust.  Certificates are signed by
"Cockroach CA" which is not a generally-trusted CA, but this is ok.

### Accessing a CockroachDB SQL terminal

To interact with the CockroachDB database directly via SQL terminal:

```
kubectl \
  --context $CLUSTER_CONTEXT exec --namespace $NAMESPACE -it \
  cockroachdb-0 -- \
  ./cockroach sql --certs-dir=cockroach-certs/
```

### Using the CockroachDB web UI

The CockroachDB web UI is not exposed publicly, but you can forward a port to
your local machine using kubectl:

#### Create a user account

Pick a username and create an account:

Access the [CockrachDB SQL terminal](#Accessing-a-CockroachDB-SQL-terminal) then create user with sql command

    root@:26257/rid> CREATE USER foo WITH PASSWORD 'foobar';

#### Access the web UI

    kubectl -n $NAMESPACE port-forward cockroachdb-0 8080

Then go to https://localhost:8080. You'll have to ignore the HTTPS certificate
warning.

## Upgrading Database Schemas

All schemas-related files are in `deploy/db_schemas` directory.  Any changes you
wish to make to the database schema should be done in their respective database
folders.  The files are applied in sequential numeric steps from the current
version M to the desired version N.

For the first-ever run during the CRDB cluster initialization, the db-manager
will run once to bootstrap and bring the database up to date.  To upgrade
existing clusters you will need to:

### If performing this operation on the original cluster
1. Update the `desired_xyz_db_version` field in `main.jsonnet`
2. Delete the existing db-manager job in your k8s cluster
3. Redeploy the newly configured db-manager with `tk apply -t job/<xyz-schema-manager>`. It should automatically up/down grade your database schema to your desired version.

### If performing this operation on any other cluster

1. Create `workspace/$CLUSTER_CONTEXT_schema_manager` in this (build) directory.

1.  From this (build) working directory,
    `cp -r deploy/examples/schema_manager/* workspace/$CLUSTER_CONTEXT_schema_manager`.

1.  Edit `workspace/$CLUSTER_CONTEXT_schema_manager/main.jsonnet` and replace all `VAR_*`
    instances with appropriate values where applicable as explained in the above section.

1.  Run `tk apply workspace/$CLUSTER_CONTEXT_schema_manager`

### Garbadge collector job ###
Only since commit [c789b2b](https://github.com/interuss/dss/commit/c789b2b4a9fa5fb651d202da0a3abc02a03c15d2) on Aug 25, 2020 will the DSS enable automatic garbage collection of records by tracking which DSS instance is responsible for garbage collection of the record. Expired records added with a DSS deployment running code earlier than this must be manually removed.

The Garbage collector job runs every 30 minute to delete records in RID tables that records' endtime is 30 minutes less than current time. If the event takes a long time and takes longer than 30 minutes (previous job is still running), the job will skip a run until the previous job completes.
