# Introduction

This section describes how to deploy a DSS instance on Kubernetes.

## Deployment Options

The DSS can be deployed on various platforms. Choose the method that best suits your needs:

| Platform | Tools | Description |
| :--- | :--- | :--- |
| **Amazon Web Services** | Terraform | [Deploy on AWS using Terraform](aws.md) to provision EKS and required resources. |
| **Google Cloud Platform** | Terraform | [Deploy on GCP using Terraform](google.md) to provision GKE and required resources. |
| **Google Cloud Platform** | Manual | [Deploy on GCP manually](google-manual.md) without Terraform. |
| **Locally** | Minikube | [Deploy locally using Minikube](minikube.md) for development and testing. |


## Glossary

- DSS Region - A region in which a single, unified airspace representation is
  presented by one or more interoperable DSS instances, each instance typically
  operated by a separate organization.  A specific environment (for example,
  "production" or "staging") in a particular DSS Region is called a "pool".
- DSS instance - a single logical replica in a DSS pool.


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
- If using CockroachDB as the datastore, 
  [install CockroachDB](https://www.cockroachlabs.com/get-cockroachdb/) to
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

1. Use the [`build.sh` script](https://github.com/interuss/dss/blob/master/build/build.sh) in this directory to build and push
   an image tagged with the current date and git commit hash.

1. Note the VAR_* value printed at the end of the script.

### Access to private repository

See the description of `VAR_DOCKER_IMAGE_PULL_SECRET` to configure authentication [on the manual step by step guide](google-manual.md).

### Verify signature of prebuilt InterUSS Docker images

The prebuilt docker images are signed using [sigstore](https://www.sigstore.dev/).
The identity of the CI workflow, attested by GitHub, is used so sign the images.

The signature may be verified by using [cosign](https://github.com/sigstore/cosign):
```shell
docker pull "docker.io/interuss/dss:latest"
cosign verify "docker.io/interuss/dss:latest" \
  --certificate-identity-regexp="https://github.com/interuss/dss/.github/workflows/dss-publish.yml@refs/*" \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com"
```

Adapt the version specified if required.