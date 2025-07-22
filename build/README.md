# Deploying a DSS instance

## Deployment options

This document describes how to deploy a production-style DSS instance to
interoperate with other DSS instances in a DSS pool.

To run a local DSS instance for testing, evaluation, or development, see
[dev/standalone_instance.md](dev/standalone_instance.md).

To create a local DSS instance with multi-node CRDB cluster, see [dev/mutli_node_local_dss.md](dev/mutli_node_local_dss.md).

To create or join a pool consisting of multiple interoperable DSS instances, see
[information on pooling](../deploy/operations/pooling.md).

## Glossary

- DSS Region - A region in which a single, unified airspace representation is
  presented by one or more interoperable DSS instances, each instance typically
  operated by a separate organization.  A specific environment (for example,
  "production" or "staging") in a particular DSS Region is called a "pool".
- DSS instance - a single logical replica in a DSS pool.

## Preface

This doc describes a procedure for deploying the DSS and its dependencies
(namely CockroachDB) via Kubernetes. The use of Kubernetes is not a requirement,
and a DSS instance can join a CRDB cluster constituting a DSS pool as long as it
meets the [CockroachDB requirements below](#cockroachdb-requirements).

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

The application logic of the DSS is located in core-service which is provided in
a Docker image which is built locally and then pushed to a Docker registry of
your choice.  All major cloud providers have a docker registry service, or you
can set up your own.

To use the prebuilt InterUSS Docker images (without building them yourself), use
`docker.io/interuss/dss` for `VAR_DOCKER_IMAGE_NAME`.

To build these images (and, optionally, push them to a docker registry):

1. Set the environment variable `DOCKER_URL` to your docker registry url
endpoint.

    -   For Google Cloud, `DOCKER_URL` should be set similarly to as described
        [here](https://cloud.google.com/container-registry/docs/pushing-and-pulling#tag_the_local_image_with_the_registry_name),
        like `gcr.io/your-project-id` (do not include the image name;
        it will be appended by the build script)

    -   For Amazon Web Services, `DOCKER_URL` should be set similarly to as described
        [here](https://docs.aws.amazon.com/AmazonECR/latest/userguide/docker-push-ecr-image.html),
        like `${aws_account_id}.dkr.ecr.${region}.amazonaws.com/` (do not include the image name;
        it will be appended by the build script)

1. Ensure you are logged into your docker registry service.

    -   For Google Cloud,
        [these](https://cloud.google.com/container-registry/docs/advanced-authentication#gcloud-helper)
        are the recommended instructions (`gcloud auth configure-docker`).
        Ensure that
        [appropriate permissions are enabled](https://cloud.google.com/container-registry/docs/access-control).

    -   For Amazon Web Services, create a private repository by following the instructions
        [here](https://docs.aws.amazon.com/AmazonECR/latest/userguide/repository-create.html), then login
        as described [here](https://docs.aws.amazon.com/AmazonECR/latest/userguide/docker-push-ecr-image.html).

1. Use the [`build.sh` script](./build.sh) in this directory to build and push
   an image tagged with the current date and git commit hash.

1. Note the VAR_* value printed at the end of the script.

### Access to private repository

See below the description of `VAR_DOCKER_IMAGE_PULL_SECRET` to configure authentication.

## Deploying a DSS instance via Kubernetes

This section discusses deploying a Kubernetes service manually, although you can deploy
a DSS instance however you like as long as it meets the CockroachDB requirements
above. You can do this on any supported
[cloud provider](https://kubernetes.io/docs/concepts/cluster-administration/cloud-providers/)
or even on your own infrastructure. Consult the Kubernetes documentation for
your chosen provider.

To instead deploy infrastructure using terraform, see the [terraform infrastructure deployment page](../deploy/infrastructure/README.md).

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

    It may be useful to create a `login.sh` file with content like that shown
    below and `source login.sh` when working with this cluster.

    GCP:
    ```shell
    #!/bin/bash

    export CLUSTER_NAME=<your cluster name>
    export REGION=<GCP region in which your cluster resides>
    gcloud config set project <your GCP project name>
    gcloud config set compute/zone $REGION-a
    gcloud container clusters get-credentials $CLUSTER_NAME
    export CLUSTER_CONTEXT=$(kubectl config current-context)
    export NAMESPACE=default
    export DOCKER_URL=docker.io/interuss
    echo "Current CLUSTER_CONTEXT is $CLUSTER_CONTEXT
    ```

1.  Create static IP addresses: one for the Core Service ingress, and one
    for each CockroachDB node if you want to be able to interact with other
    DSS instances.

    -  If using Google Cloud, the Core Service ingress needs to be created as
       a "Global" IP address, but the CRDB ingresses as "Regional" IP addresses.
       IPv4 is recommended as IPv6 has not yet been tested.  Follow
       [these instructions](https://cloud.google.com/compute/docs/ip-addresses/reserve-static-external-ip-address#reserve_new_static)
       to reserve the static IP addresses.  Specifically (replacing
       CLUSTER_NAME as appropriate since static IP addresses are defined at
       the project level rather than the cluster level), e.g.:

         -  `gcloud compute addresses create ${CLUSTER_NAME}-backend --global --ip-version IPV4`
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

1.  (Only if you use CockroachDB) Use [`make-certs.py` script](./make-certs.py) to create certificates for
    the CockroachDB nodes in this DSS instance:

        ./make-certs.py --cluster $CLUSTER_CONTEXT --namespace $NAMESPACE
            [--node-address <ADDRESS> <ADDRESS> <ADDRESS> ...]
            [--ca-cert-to-join <CA_CERT_FILE>]

    1.  `$CLUSTER_CONTEXT` is the name of the cluster (see step 2 above).

    1.  `$NAMESPACE` is the namespace for this DSS instance (see step 3 above).

    1.  `Each ADDRESS` is the DNS entry for a CockroachDB node that will use the
        certificates generated by this command.  This is usually just the nodes
        constituting this DSS instance, though if you maintain multiple DSS
        instances in a single pool, the separate instances may share
        certificates.  Note that `--node-address` must include all the hostnames
        and/or IP addresses that other CockroachDB nodes will use to connect to
        your nodes (the nodes using these certificates). Wildcard notation is
        supported, so you can use `*.<subdomain>.<domain>.com>`.  If following
        the recommendations above, use a single ADDRESS similar to
        `*.db.yourdeployment.yourdomain.com`.  The ADDRESS entries should be
        separated by spaces.

    1.  If you are pooling with existing DSS instance(s) you need their CA
        public cert (ca.crt), which will be concatenated with yours. Set
        `--ca-cert-to-join` to a `ca.crt` file.  Reach out to existing operators
        to request their public cert.  If not joining an existing pool, omit
        this argument.

    1.  Note: If you are creating multiple DSS instances at once, and joining
        them together you likely want to copy the nth instance's `ca.crt` into
        the rest of the instances, such that ca.crt is the same across all
        instances.

1.  (Only if you use Yugabyte) Use [`dss-certs.py` script](../deploy/operations/certificates-management/README.md) to create certificates for the Yugabyte nodes in this DSS instance.

1.  If joining an existing DSS pool, share ca.crt with the DSS instance(s) you
    are trying to join, and have them apply the new ca.crt, which now contains
    both your instance's and the original instance's public certs, to enable
    secure bi-directional communication.  Each original DSS instance, upon
    receipt of the combined ca.crt from the joining instance, should perform the
    actions below.  While they are performing those actions, you may continue
    with the instructions.

    1. If you use CockroachDB:

        1. Overwrite its existing ca.crt with the new ca.crt provided by the DSS
        instance joining the pool.
        1. Upload the new ca.crt to its cluster using
        `./apply-certs.sh $CLUSTER_CONTEXT $NAMESPACE`
        1. Restart their CockroachDB pods to recognize the updated ca.crt:
        `kubectl rollout restart statefulset/cockroachdb --namespace $NAMESPACE`
        1. Inform you when their CockroachDB pods have finished restarting
        (typically around 10 minutes)

    1. If you use Yugabyte

        1. Share your CA with `./dss-certs.py get-ca`
        1. Add others CAs of the pool with `./dss-certs.py add-pool-ca`
        1. Upload the new CAs to its cluster using
        `./dss-certs.py apply`
        1. Restart their Yugabyte pods to recognize the updated ca.crt:
        `kubectl rollout restart statefulset/yb-master --namespace $NAMESPACE`
        `kubectl rollout restart statefulset/yb-tserver --namespace $NAMESPACE`
        1. Inform you when their Yugabyte pods have finished restarting
        (typically around 10 minutes)

1.  Ensure the Docker images are built according to the instructions in the
    [previous section](#docker-images).

1.  From this working directory,
    `cp -r ../deploy/services/tanka/examples/minimum/* workspace/$CLUSTER_CONTEXT`.  Note that
    the `workspace/$CLUSTER_CONTEXT` folder should have already been created
    by the `make-certs.py` script.
    Replace the imports at the top of `main.jsonnet` to correctly locate the files:
    ```
    local dss = import '../../../deploy/services/tanka/dss.libsonnet';
    local metadataBase = import '../../../deploy/services/tanka/metadata_base.libsonnet';
    ```

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

    1.  `VAR_LOCALITY`: Unique name for your DSS instance.  Currently, we
        recommend "<ORG_NAME>_<CLUSTER_NAME>", and the `=` character is not
        allowed.  However, any unique (among all other participating DSS
        instances) value is acceptable.

    1.  `VAR_DB_HOSTNAME_SUFFIX`: The domain name suffix shared by all of your
        CockroachDB nodes.  For instance, if your CRDB nodes were addressable at
        `0.db.example.com`, `1.db.example.com`, and `2.db.example.com`, then
        VAR_DB_HOSTNAME_SUFFIX would be `db.example.com`.

    1.  `VAR_DATASTORE`: Datastore to use. Can be set to 'cockroachdb' or 'yugabyte'.

    1.  `VAR_CRDB_DOCKER_IMAGE_NAME`: Docker image of cockroach db pods. Until
        DSS v0.16, the recommended CockroachDB image name is `cockroachdb/cockroach:v21.2.7`.
        From DSS v0.17, the recommended CockroachDB version is `cockroachdb/cockroach:v24.1.3`.

    1.  `VAR_CRDB_NODE_IPn`: IP address (**numeric**) of nth CRDB node (add more
        entries if you have more than 3 CRDB nodes).  Example: `1.1.1.1`

    1.  `VAR_SHOULD_INIT`: Set to `false` if joining an existing pool, `true`
        if creating the first DSS instance for a pool.  When set `true`, this
        can initialize the data directories on your cluster, and prevent you
        from joining an existing pool.

    1.  `VAR_EXTERNAL_CRDB_NODEn`: Fully-qualified domain name of existing CRDB
        nodes if you are joining an existing pool.  If more than three are
        available, add additional entries.  If not joining an existing pool,
        comment out this `JoinExisting:` line.

        - You should supply a minimum of 3 seed nodes to every CockroachDB node.
          These 3 nodes should be the same for every node (ie: every node points
          to node 0, 1, and 2). For external DSS instances you should point to a
          minimum of 3, or you can use a loadbalanced hostname or IP address of
          other DSS instances. You should do this for every DSS instance in the
          pool, including newly joined instances. See CockroachDB's note on the
          [join flag](https://www.cockroachlabs.com/docs/stable/start-a-node.html#flags).

    1.  `VAR_YUGABYTE_DOCKER_IMAGE_NAME`: Docker image of Yugabyte db pods.
        Shall be set to at least `yugabytedb/yugabyte:2.25.1.0-b381`

    1.  `VAR_YUGABYTE_MASTER_IPn`: IP address (**numeric**) of nth Yugabyte
        master node (add more entries if you have more than 3 nodes).
        Example: `1.1.1.1`

    1.  `VAR_YUGABYTE_TSERVER_IPn`: IP address (**numeric**) of nth Yugabyte
        tserver node (add more entries if you have more than 3 nodes).
        Example: `1.1.1.1`

    1.  `VAR_YUGABYTE_MASTER_ADDRESSn`: List of addresses of Yugabyte master
        nodes in the DSS pool. Must be accessible from all master/tserver nodes
        and identical in a cluster. Example: `["0.master.db.uss1.example.com", "1.master.db.uss1.example.com", "3.master.db.uss1.example.com", "0.master.db.uss2.example.com", "1.master.db.uss2.example.com", "3.master.db.uss2.example.com"]`
        You may remove this setting if you only have a simple 3-nodes local cluster.

    1.  `VAR_YUGABYTE_MASTER_RPC_BIND_ADDRESSES`: Bind address for yugabyte
        master node. May use `${HOSTNAME}`, `${NAMESPACE}` or `${HOSTNAMENO}`
        to use respectively hostname, namespace or number of the node.
        Example: `${HOSTNAMENO}.master.db.uss1.example.com`
        You may remove this setting if you only have a simple 3-nodes local cluster.

    1.  `VAR_YUGABYTE_MASTER_BROADCAST_ADDRESSES`: Broadcast address for yugabyte
        master node. May use `${HOSTNAME}`, `${NAMESPACE}` or `${HOSTNAMENO}`
        to use respectively hostname, namespace or number of the node.
        Example: `${HOSTNAMENO}.master.db.uss1.example.com:7100`
        You may remove this setting if you only have a simple 3-nodes local cluster.

    1.  `VAR_YUGABYTE_TSERVER_RPC_BIND_ADDRESSES`: Bind address for yugabyte
        tserver node. May use `${HOSTNAME}`, `${NAMESPACE}` or `${HOSTNAMENO}`
        to use respectively hostname, namespace or number of the node.
        Example: `${HOSTNAMENO}.tserver.db.uss1.example.com`
        You may remove this setting if you only have a simple 3-nodes local cluster.

    1.  `VAR_YUGABYTE_TSERVER_BROADCAST_ADDRESSES`: Broadcast address for yugabyte
        tserver node. May use `${HOSTNAME}`, `${NAMESPACE}` or `${HOSTNAMENO}`
        to use respectively hostname, namespace or number of the node.
        Example: `${HOSTNAMENO}.tserver.db.uss1.example.com:9100`
        You may remove this setting if you only have a simple 3-nodes local cluster.

    1.  `VAR_YUGABYTE_FIX_27367_ISSUE`: Fix issue [27367](https://github.com/yugabyte/yugabyte-db/issues/27367)
        To make the fix working, RPC bind and broadcast addresses must be set to
        the same, public value on where the master / tserver node is accessible.

    1.  `VAR_YUGABYTE_LIGHT_RESOURCES`: Use light resources in term of CPU/Memory
        for Yugabyte nodes. You may use that for development purposes, to deploy
        a Yugabyte in a small cluster to save costs and resources.

    1.  `VAR_YUGABYTE_PLACEMENT_CLOUD`: Yugabyte placement's cloud value, for
        master and tserver nodes.
        Example: `cloud-1`

    1.  `VAR_YUGABYTE_PLACEMENT_REGION`: Yugabyte placement's region value, for
        master and tserver nodes.
        Example: `uss-1`

    1.  `VAR_YUGABYTE_PLACEMENT_ZONE`: Yugabyte placement's zone value, for
        master and tserver nodes.
        Example: `zone-1`

    1.  `VAR_STORAGE_CLASS`: Kubernetes Storage Class to use for CockroachDB,
        Yugabyte and Prometheus volumes. You can check your cluster's possible
        values with `kubectl get storageclass`. If you're not sure, each cloud
        provider has some default storage classes that should work:
          - Google Cloud: `standard`
          - Azure: `default`
          - AWS: `gp2`

    1.  `VAR_INGRESS_NAME`: If using Google Kubernetes Engine, set this to the
        the name of the core-service static IP address created above (e.g.,
        `CLUSTER_NAME-backend`).

    1.  `VAR_DOCKER_IMAGE_NAME`: Full name of the docker image built in the
        section above.  `build.sh` prints this name as the last thing it does
        when run with `DOCKER_URL` set.  It should look something like
        `gcr.io/your-project-id/dss:2020-07-01-46cae72cf` if you built the image
        yourself, or `docker.io/interuss/dss` if using the InterUSS image
        without `build.sh`.

        -   Note that `VAR_DOCKER_IMAGE_NAME` is used in two places.

    1.  `VAR_DOCKER_IMAGE_PULL_SECRET`: Secret name of the credentials to access
        the image registry. If the image specified in VAR_DOCKER_IMAGE_NAME does not require
        authentication to be pulled, then do not populate this instance and do not uncomment
        the line containing it. You can use the following command to store the credentials
        as kubernetes secret:

        > kubectl create secret -n VAR_NAMESPACE docker-registry VAR_DOCKER_IMAGE_PULL_SECRET \
            --docker-server=DOCKER_REGISTRY_SERVER \
            --docker-username=DOCKER_USER \
            --docker-password=DOCKER_PASSWORD \
            --docker-email=DOCKER_EMAIL

        For docker hub private repository, use `docker.io` as `DOCKER_REGISTRY_SERVER` and an
        [access token](https://hub.docker.com/settings/security) as `DOCKER_PASSWORD`.

    1.  `VAR_APP_HOSTNAME`: Fully-qualified domain name of your Core Service
        ingress endpoint.  For example, `dss.example.com`.

    1.  `VAR_PUBLIC_ENDPOINT`: URL to publicly access your Core Service
        ingress endpoint.  For example, `https://dss.example.com`. Only for versions >=0.21.

    1.  `VAR_PUBLIC_KEY_PEM_PATH`: If providing a .pem file directly as the
        public key to validate incoming access tokens, specify the name of this
        .pem file here as `/jwt-public-certs/YOUR-KEY-NAME.pem` replacing
        YOUR-KEY-NAME as appropriate.  For instance, if using the provided
        [`us-demo.pem`](./jwt-public-certs/us-demo.pem), use the path
        `/jwt-public-certs/us-demo.pem`.  Note that your .pem file must have
        been copied into [`jwt-public-certs`](./jwt-public-certs) in an earlier
        step, or mounted at runtime using a volume.

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

    -   If you are only turning up a single DSS instance for development, you
        may optionally change `single_cluster` to `true`.

    1.  `VAR_SSL_POLICY`: When deploying on Google Cloud, a [ssl policy](https://cloud.google.com/load-balancing/docs/ssl-policies-concepts)
        can be applied to the DSS Ingress. This can be used to secure the TLS connection.
        Follow the [instructions](https://cloud.google.com/load-balancing/docs/use-ssl-policies) to create the Global SSL Policy and
        replace VAR_SSL_POLICY variable with its name. `RESTRICTED` profile is recommended.
        Leave it empty if not applicable.

    1.  `VAR_ENABLE_SCHEMA_MANAGER`: Set this to true to enable the schema manager jobs.
        It is required to perform schema upgrades. Note that it is automatically enabled when `VAR_SHOULD_INIT` is true.


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

    - If you are joining an existing pool, do not execute this command until the
      the existing DSS instances all confirm that their CockroachDB pods have
      finished their rolling restarts.

1.  Wait for services to initialize.  Verify that basic services are functioning
    by navigating to https://your-domain.example.com/healthy.

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

1.  If joining an existing pool, share your CRDB node addresses with the
    operators of the existing DSS instances.  They will add these node addresses
    to JoinExisting where `VAR_CRDB_EXTERNAL_NODEn` is indicated in the minimum
    example, and then update their deployment:

    `tk apply workspace/$CLUSTER_CONTEXT`

## Pooling

See [the pooling documentation](../deploy/operations/pooling.md).

## Tools

See [operations monitoring documentation](../deploy/operations/monitoring.md).

## Troubleshooting

See [Troubleshooting in `deploy/operations`](../deploy/operations/troubleshooting.md).

## Upgrading Database Schemas

All schemas-related files are in `db_schemas` directory.  Any changes you
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
    `cp -r ../deploy/services/tanka/examples/schema_manager/* workspace/$CLUSTER_CONTEXT_schema_manager`.

1.  Edit `workspace/$CLUSTER_CONTEXT_schema_manager/main.jsonnet` and replace all `VAR_*`
    instances with appropriate values where applicable as explained in the above section.

1.  Run `tk apply workspace/$CLUSTER_CONTEXT_schema_manager`

### Garbage collector job
Only since commit [c789b2b](https://github.com/interuss/dss/commit/c789b2b4a9fa5fb651d202da0a3abc02a03c15d2) on Aug 25, 2020 will the DSS enable automatic garbage collection of records by tracking which DSS instance is responsible for garbage collection of the record. Expired records added with a DSS deployment running code earlier than this must be manually removed.

The Garbage collector job runs every 30 minute to delete records in RID tables that records' endtime is 30 minutes less than current time. If the event takes a long time and takes longer than 30 minutes (previous job is still running), the job will skip a run until the previous job completes.
