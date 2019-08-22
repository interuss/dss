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


## Creating a new Kubernetes cluster

You need to create a Kubernetes cluster to run the cockroachdb instance and the
gRPC and HTTP servers.  You can do this on any supported
[cloud provider](https://kubernetes.io/docs/concepts/cluster-administration/cloud-providers/)
or even on your own infrastructure.  Instructions for GCP are given below,
consult the Kubernetes documentation for other providers.

### Google Container Engine

Create a new cluster in the given zone:

    gcloud --project $CLOUD_PROJECT container clusters create $CLUSTER_NAME --zone <ZONE>

Fetch credentials for the cluster.  This populates your \~/.kube/config file
and makes all future kubecfg commands target this cluster.

    gcloud --project $CLOUD_PROJECT container clusters get-credentials $CLUSTER_NAME


## Creating a new cockroachdb cluster

1.  Use the `make-certs.py` script in this directory to create certificates for
    the new cockroachdb cluster:

        ./make-certs.py $NAMESPACE \
            [--node-address <ADDRESS>]
            [--node-ca-cert <FILENAME>]

    *   If you are joining existing clusters, make sure to provide their public
        CA certificates with --node-ca-cert, and their addresses with
        --node-address.

1.  Use the `apply-certs.sh` script in this directory to create secrets on the
    Kubernetes cluster containing the certificates and keys generated in the
    previous step.

        ./apply-certs.sh $NAMESPACE

1.  Copy `values.yaml.template` to `values.yaml` and fill in the required fields
    at the top.
1.  Run `helm template . > cockroachdb.yaml` to render the YAML.
1.  Run `kubectl apply -f cockroachdb.yaml` to apply it to the cluster.
1.  Use the `./expose.sh` script in this directory to create an external IP for
    each pod:

        ./expose.sh $NAMESPACE

1.  Find out the external IP addresses that were just created:

        kubectl get svc --namespace $NAMESPACE

1.  Make these IPs static.  The way you do this depends on your cloud provider.
1.  Add the external IP addresses for the `crdb-node-*` entries to the `ips`
    list in the values.yaml file and re-run the `helm template` command.
1.  Re-run the `./make-certs.py` script with the external IP addresses added to
    `--node-address` flags, then re-run `./apply-certs.sh`.
1.  Re-run the `kubectl apply` command to update production.  This should
    restart the pods automatically, but in case it doesn't you can restart them
    manually:

        kubectl rollout restart --namespace $NAMESPACE statefulset cockroachdb
