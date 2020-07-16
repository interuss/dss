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

This doc provides a well-lit path for deploying the DSS and its dependencies
(namely CockroachDB) on Kubernetes. The use of Kubernetes is not a requirement,
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
  - Confirm sucessful installation with `kubectl version --client` (should
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
        are the recommended instructions (`gcloud auth configure-docker`).
        Ensure that
        [appropriate permissions are enabled](https://cloud.google.com/container-registry/docs/access-control).

1. Use the [`build.sh` script](./build.sh) in this directory to build and push
   an image tagged with the current date and git commit hash.
        
1. Note the two VAR_* values printed at the end of the script.

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

If you can augment this documentation with specifics for another cloud provider,
a PR to that effect would be greatly appreciated.

1.  Create a new Kubernetes cluster. We recommend a new cluster for each DSS
    instance.  A reasonable cluster name might be `dss-us-prod-e4a` (where `e4a`
    is a zone identifier abbreviation), `dss-ca-staging`,
    `dss-mx-integration-sae1a`, etc.  The name of this cluster will be combined
    with other information by Kubernetes to generate a longer cluster context ID
    that will be referred to as CLUSTER_CONTEXT for the remainder of this
    document.
    
    -  On Google Cloud, the recommended procedure to create a cluster is:
       -  In Google Cloud Platform, go to the Kubernetes Engine page and under
          Clusters click Create cluster.
       -  Name the cluster appropriately; e.g., `dss-us-prod`
       -  Select Zonal and [a compute-zone appropriate to your
          geography](https://cloud.google.com/compute/docs/regions-zones#available)
       -  For the "default-pool" node pool:
          - Enter 2 for number of nodes.
          - Check "Enable autoscaling" and enter a range of 2-10 nodes.
       -  In the "Nodes" bullet under "default-pool", select N2 series and
          n2-standard-8 for machine type.
       -  In the "Networking" bullet under "Clusters", ensure "Enable [VPC
          -native traffic](https://cloud.google.com/kubernetes-engine/docs/how-to/alias-ips)"
          is checked.

1.  Make sure correct cluster context is selected by printing the context
    name to the console: `kubectl config current-context`
    
    -  Record this value and use it for `CLUSTER_CONTEXT` below.
    
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
       
         -  `gcloud compute addresses create CLUSTER_NAME-gateway --global --ip-version IPV4`
         -  `gcloud compute addresses create CLUSTER_NAME-crdb-0 --global --ip-version IPV4`
         -  `gcloud compute addresses create CLUSTER_NAME-crdb-1 --global --ip-version IPV4`
         -  `gcloud compute addresses create CLUSTER_NAME-crdb-2 --global --ip-version IPV4`

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
      CockroachDB pods to pick up the new certificates:
      
      `kubectl rollout restart statefulset/cockroachdb --namespace <NAMESPACE>`

1.  Ensure the Docker images are built according to the instructions in the
    [previous section](#docker-images).

1.  From this working directory,
    `cp -r deploy/examples/minimum/* workspace/<CLUSTER_CONTEXT>`.  Note that
    the `workspace/<CLUSTER_CONTEXT>` folder should have already been created
    by the `make_certs.py` script.

1.  If providing a .pem file directly as the public key to validate incoming
    access tokens, copy it to [dss/build/jwt-public-certs](./jwt-public-certs).
    Public key specification by JWKS is preferred; if using the JWKS approach to
    to specify the public key, skip this step.

1.  Edit `workspace/<CLUSTER_CONTEXT>/main.jsonnet` and replace all `VAR_*`
    instances with appropriate values:
   
    1.  `VAR_NAMESPACE`: Same <NAMESPACE> used in the make-certs.py (and
        apply-certs.sh) scripts.
   
    1.  `VAR_CLUSTER_CONTEXT`: Same <CLUSTER_CONTEXT> name of the cluster used in
        the `make-certs.py` and `apply-certs.sh` scripts.

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
        if creating the first DSS instance for a Region.  When set `true`, this
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

1.  Edit workspace/<CLUSTER_CONTEXT>/spec.json and replace all VAR_*
    instances with appropriate values:
   
    1.  VAR_API_SERVER: Determine this value with the command:
    
        `kubectl config view -o jsonpath="{.clusters[?(@.name==\"<CLUSTER_CONTEXT>\")].cluster.server}"`
        
        - Note that `<CLUSTER_CONTEXT>` should be replaced with your actual
         `CLUSTER_CONTEXT` value prior to executing the above command.
   
    1.  VAR_NAMESPACE: See previous section.

1.  Use the [`apply-certs.sh` script](apply-certs.sh) to create secrets on the
    Kubernetes cluster containing the certificates and keys generated in the
    previous step.

        ./apply-certs.sh <CLUSTER_CONTEXT> <NAMESPACE>

1.  Run `tk apply workspace/<CLUSTER_CONTEXT>` to apply it to the
    cluster.

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
    example, and then perform a rolling restart of their CockroachDB pods:
    
    `kubectl rollout restart statefulset/cockroachdb --namespace <NAMESPACE>`

## CockroachDB requirements
These requirements must be met by every DSS instance joining a DSS Region.  The
Kubernetes deployment instructions above produce a system that complies with all
these requirements, so this section may be ignored if following those
instructions.

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
To enable Istio, simply set the `enable_istio` field in your metadata tuple to
`true`, then run `tk apply ...` as you would normally. 

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

## Upgrading Database Schemas

All schemas-related files are in `deploy/db-schemas` directory.  Any changes you
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
1. Upload the `db-schema/<database>` directory as a ConfigMap with
   `kubectl create configmap -n <namespace> --from-file <path to schemas>`
1. Prepare a Yaml file to deploy a K8s Job that will run the `db-manager` binary
   with the following flags:
   ```
   --cockroach_host cockroachdb-balanced.<namespace>
   --cockroach_port 26257
   --cockroach_ssl_mode: 'verify-full'
   --cockroach_user: 'root'
   --cockroach_ssl_dir: <path to the mounted cockroach certificate secrets>
   --db_version: <desired db version>
   --schemas_dir: <path to the mounted schemas configmap>
    ```
