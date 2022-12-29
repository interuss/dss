# DSS Infrastructure Examples

This folder contains deployment examples for various environments:
- Cloud Provider:
  - terraform-google-dss: Google Cloud Engine deployment

## Infrastructure

### Prerequisites
Download & install the following tools to your workstation:

1. Install [terraform](https://developer.hashicorp.com/terraform/downloads).
2. Install provider specific tools:
   1. [Google Cloud Engine](./README.md#google-cloud-engine)
3. Install tools from [Prerequisites](../../../build/README.md)

#### Google Cloud Engine

1. Install and initialize [Google Cloud CLI](https://cloud.google.com/sdk/docs/install-sdk).
   1. Confirm successful installation with `gcloud version`.
2. Check that the DSS project is correctly selected: gcloud config list project
    1. Set another one if needed using: `gcloud config set project $GOOGLE_PROJECT_NAME`
3. Enable the following API using [Google Cloud CLI](https://cloud.google.com/endpoints/docs/openapi/enable-api#gcloud):
    1. `container.googleapis.com`
    2. If you want to manage DNS entries with terraform: `dns.googleapis.com`
4. Install the auth plugin to connect to kubernetes: `gcloud components install gke-gcloud-auth-plugin`
5. Run `gcloud auth application-default login` to generate credentials to call Google Cloud Platform APIs.

### Deployment of the Kubernetes cluster

1. Copy or edit in place an example folder to `/deploy/infrastructure/personal/`. (Note that the modules can be added to existing projects) 
2. Edit `terraform.tfvars` and set the variables according to your environment.
3. Initialize terraform: `terraform init`.
4. Run `terraform plan` to check that the configuration is valid. It will display the resources which will be provisioned.
5. Run `terraform apply` to deploy the cluster. (This operation may take up to 15 min.)

#### Note on DNS

DNS entries can be either managed manually or handled by terraform depending on the cloud provider.
See [DNS](DNS.md) for details.

## Deployment of the DSS services

During the successful run, the terraform job has created a new [workspace](../../../build/workspace/) 
for the new cluster.

It contains scripts to operate the cluster and setup the services.

1. Go to `/build/workspace/${CLUSTER_CONTEXT}`.
2. Run `./get_credentials.sh` to login to kubernetes. You can now access the cluster with `kubectl`.
3. Generate the certificates `./make-certs.sh`. Follow script instructions if you are not initializing the cluster. 
4. Deploy the certificates `./apply-certs.sh`.
5. Run `tk apply .` to deploy the services to kubernetes. (This may take up to 30 min)
6. Wait for services to initialize. Verify that basic services are functioning by navigating to https://your-gateway-domain.com/healthy.

   - On Google Cloud, the highest-latency operation is provisioning of the HTTPS certificate which generally takes 10-45 minutes. To track this progress:
     - Go to the "Services & Ingress" left-side tab from the Kubernetes Engine page.
     - Click on the https-ingress item (filter by just the cluster of interest if you have multiple clusters in your project).
     - Under the "Ingress" section for Details, click on the link corresponding with "Load balancer".
     - Under Frontend for Details, the Certificate column for HTTPS protocol will have an icon next to it which will change to a green checkmark when provisioning is complete.
     - Click on the certificate link to see provisioning progress.
     - If everything indicates OK and you still receive a cipher mismatch error message when attempting to visit /healthy, wait an additional 5 minutes before attempting to troubleshoot further.

## Clean up

To delete all resources, run `terraform destroy`. Note that this operation can't be reverted and all data will be lost.