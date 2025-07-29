# Prerequisites

## Terraform
Install [Terraform](https://developer.hashicorp.com/terraform/downloads).


## Kubernetes cluster configuration
=== "Tanka"
    - [Install tanka](https://tanka.dev/install)
    - On Linux, after downloading the binary per instructions, run `sudo chmod +x /usr/local/bin/tk`
    - Confirm successful installation with `tk --version`
    - Optionally install [Jsonnet](https://github.com/google/jsonnet) if editing the jsonnet templates.

=== "Helm"
    TBD


## Database configuration
=== "CockroachDB"
    TBD

=== "Yugabyte"
    [Install CockroachDB](https://www.cockroachlabs.com/get-cockroachdb/) to generate CockroachDB certificates.
    
    - These instructions assume CockroachDB Core.
    - You may need to run `sudo chmod +x /usr/local/bin/cockroach` after completing the installation instructions.
    - Confirm successful installation with `cockroach version`


## Cloud CLI client
=== "AWS"
    1. Install and initialize [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html#getting-started-install-instructions).
    1. Confirm successful installation with `aws --version`.
    2. If you don't have an account, sign-up: https://aws.amazon.com/free/
    3. Configure terraform to connect to AWS using your account.
    1. We recommend to create an AWS_PROFILE using for instance `aws configure --profile aws-interuss-dss`
    Before running `terraform` commands, run once in your shell: `export AWS_PROFILE=aws-interuss-dss`
    Other methods are described here: https://registry.terraform.io/providers/hashicorp/aws/latest/docs#authentication-and-configuration

=== "GCP"
    1. Install and initialize [Google Cloud CLI](https://cloud.google.com/sdk/docs/install-sdk).
    1. Confirm successful installation with `gcloud version`.
    2. Check that the DSS project is correctly selected: gcloud config list project
    1. Set another one if needed using: `gcloud config set project $GOOGLE_PROJECT_NAME`
    3. Enable the following API using [Google Cloud CLI](https://cloud.google.com/endpoints/docs/openapi/enable-api#gcloud):
    1. `compute.googleapis.com`
    2. `container.googleapis.com`
    3. If you want to manage DNS entries with terraform: `dns.googleapis.com`
    4. Install the auth plugin to connect to kubernetes: `gcloud components install gke-gcloud-auth-plugin`
    5. Run `gcloud auth application-default login` to generate credentials to call Google Cloud Platform APIs.
    1. If the result of performing the authorization indicates 404 in the browser, check whether a local dummy-oauth instance is running (using port 8085).  Stop the dummy-oauth instance if it is running.

