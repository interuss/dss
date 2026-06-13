# Deploying DSS services

## Prerequisites

Before beginning services deployment:

- Deploy appropriate [infrastructure](../infrastructure/index.md) (Kubernetes cluster is available)
- Complete appropriate [pooling configuration](../pooling/index.md)
- Download & install the following tools to your workstation:
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

## Services deployment

This section describes how to deploy services for a DSS instance once the
prerequisites have been satisified.

Depending on [infrastructure deployment](../infrastructure/index.md) method, deploy services:

- [After using terraform for infrastructure deployment](./after-terraform.md)
- [To a local Minikube cluster](./to-minikube.md)

Then, deploy any necessary [support services](./support.md).

## Docker images

The application logic of the DSS is located in core-service which is provided in
a Docker image.  To use the prebuilt InterUSS Docker images (without building
them yourself), simply use `docker.io/interuss/dss` for `VAR_DOCKER_IMAGE_NAME`
and the rest of this section may be skipped.

Instead of using the prebuilt images, you can build the image locally and push
it to a Docker registry of your choice.  All major cloud providers have a
docker registry service, or you can set up your own.

To build these images locally and, optionally, push them to a docker registry:

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

See the description of `VAR_DOCKER_IMAGE_PULL_SECRET` in TFVARS.gen.md to configure authentication when using terraform to deploy [infrastructure](../infrastructure/index.md).

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
