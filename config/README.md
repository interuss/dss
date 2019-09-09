# Multi-region cockroachdb setup

## Prerequisites

Download & install:

*   helm
*   kubectl
*   docker
*   cockroachdb
*   Google Cloud SDK (if deploying on GCP)
*   [Optional] Golang. Recommended to understand go, and the go toolchain.


## Building Docker images

The grpc-backend and http-gateway binaries are built as docker images and pushed
to a docker registry.  Which docker registry you use will depend on your
individual requirements and choice of cloud platform.

### Google Container Registry

List existing images:

    gcloud --project $CLOUD_PROJECT container images list

List the tags on an existing image:

    gcloud --project $CLOUD_PROJECT container images list-tags gcr.io/$CLOUD_PROJECT/http-gateway

Build a new image:

    docker build -f cmds/http-gateway/Dockerfile  . -t gcr.io/$CLOUD_PROJECT/http-gateway:$VERSION
    docker build -f cmds/grpc-backend/Dockerfile  . -t gcr.io/$CLOUD_PROJECT/grpc-backend:$VERSION

Push your new image to Google Container Registry:

    docker push gcr.io/$CLOUD_PROJECT/http-gateway:$VERSION
    docker push gcr.io/$CLOUD_PROJECT/grpc-backend:$VERSION

Use the `build.sh` script in this directory to build and push an image tagged
with the current date and git commit hash.


## Creating a new Kubernetes cluster

You need to create a Kubernetes cluster to run the cockroachdb instance and the
gRPC and HTTP servers.  You can do this on any supported
[cloud provider](https://kubernetes.io/docs/concepts/cluster-administration/cloud-providers/)
or even on your own infrastructure.  Instructions for GCP are given below,
consult the Kubernetes documentation for other providers.

### Google Container Engine

Create a new cluster in the given zone:

    gcloud --project $CLOUD_PROJECT container clusters create $CLUSTER_NAME --zone $ZONE

Fetch credentials for the cluster.  This populates your \~/.kube/config file
and makes all future kubecfg commands target this cluster.

    gcloud --project $CLOUD_PROJECT container clusters get-credentials $CLUSTER_NAME --zone $ZONE


## Creating a new cockroachdb cluster

1.  Create 5 static IP addresses.  How you do this depends on your cloud
    provider.  3 IPs will be used for the 3 individual cockroachdb nodes, 1 IP
    will be used to load balance amongst all the cockroachdb nodes, and 1 IP
    will be used for the HTTPS frontend.  The HTTPS frontend IP will be used on
    a Kubernetes Ingress, and if you're using Google Cloud it needs to be
    created as a "Global" IP address.

1.  Copy `values.yaml.template` to `values.yaml` and fill in the required fields
    at the top.

1.  Use the `make-certs.py` script in this directory to create certificates for
    the new cockroachdb cluster:

        ./make-certs.py $NAMESPACE \
            [--node-address <ADDRESS> ...]
            [--ca-certs-dir <CA_CERTS_DIR>]

    *   --node-addresses needs to include all the hostnames or IP addresses that
        other cockroachdb clusters will use to connect to your cluster.  It
        should include the hostnames/IP addresses of the 3 individual
        cockroachdb nodes and the 1 load balanced endpoint.
    *   If you are joining existing clusters you need their CA certificate and
        private key to sign your certificates.  This will likely mean that the
        owners of the existing cluster will need to generate certificates for
        you.  Set --ca-certs-dir to a directory containing `ca.crt` and `ca.key`
        files.

1.  Use the `apply-certs.sh` script in this directory to create secrets on the
    Kubernetes cluster containing the certificates and keys generated in the
    previous step.

        ./apply-certs.sh $NAMESPACE

1.  Run `helm template . > cockroachdb.yaml` to render the YAML.
1.  Run `kubectl apply -f cockroachdb.yaml` to apply it to the cluster.


## Joining an existing cockroachdb cluster

Follow the steps above for creating a new cockroachdb cluster, but with the
following differences:

1.  In values.yaml, be sure to set ClusterInit to false otherwise you'll
    reinitialize and destroy the existing cluster.
1.  In values.yaml, add the host:ports of existing cockroachdb nodes to the
    JoinExisting array.  You can use the loadbalanced hostnames or IP addresses
    of other clusters (the DBBalanced hostname/IP), or you can specify each node
    individually.
1.  You must run ./make-certs.py with the --ca-certs-dir flag as described above
    to use the existing cluster's CA to sign your certificates.

## Using the cockroachdb web UI

The cockroachdb web UI is not exposed publicly, but you can forward a port to
your local machine using the kubectl command:

### Create a user account

    kubectl -n $NAMESPACE exec cockroachdb-0 -ti -- \
        ./cockroach --certs-dir ./cockroach-certs \
        user set $USERNAME --password

### Access the web UI

    kubectl -n $NAMESPACE port-forward cockroachdb-0 8080

Then go to https://localhost:8080.  You'll have to ignore the HTTPS certificate
warning.

