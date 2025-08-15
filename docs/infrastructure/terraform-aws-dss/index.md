# terraform-aws-dss

This terraform module creates a Kubernetes cluster in Amazon Web Services using the Elastic Kubernetes Service (EKS)
and generates the tanka files to deploy a DSS instance.


## Getting started

### Prerequisites

Download & install the following tools to your workstation:

1. Install [terraform](https://developer.hashicorp.com/terraform/downloads).
2. Install tools from [Prerequisites](../../build.md)
3. Install provider specific tools:
    1. [Amazon Web Services](./index.md#amazon-web-services)


#### Amazon Web Services

1. Install and initialize [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html#getting-started-install-instructions).
    1. Confirm successful installation with `aws --version`.
2. If you don't have an account, sign-up: https://aws.amazon.com/free/
3. Configure terraform to connect to AWS using your account.
   1. We recommend to create an AWS_PROFILE using for instance `aws configure --profile aws-interuss-dss`
      Before running `terraform` commands, run once in your shell: `export AWS_PROFILE=aws-interuss-dss`
      Other methods are described here: https://registry.terraform.io/providers/hashicorp/aws/latest/docs#authentication-and-configuration


### Deployment of the Kubernetes cluster

1. Create a new folder in `/deploy/infrastructure/personal/` named, for instance, `terraform-aws-dss-dev`.
2. Copy main.tf, output.tf and variables.gen.tf to the new folder.
3. Copy `terraform.dev.example.tfvars` and rename to `terraform.tfvars`
4. Check that your new directory contains the following files:
   - main.tf
   - output.tf
   - terraform.tfvars
   - variables.gen.tf
5. Set the variables in `terraform.tfvars` according to your environment. See [TFVARS.gen.md](https://github.com/interuss/dss/blob/master/deploy/infrastructure/modules/terraform-aws-dss/TFVARS.gen.md) for variables descriptions.
6. In the new directory (ie /deploy/infrastructure/personal/terraform-aws-dss-dev), initialize terraform: `terraform init`.
7. Run `terraform plan` to check that the configuration is valid. It will display the resources which will be provisioned.
8. Run `terraform apply` to deploy the cluster. (This operation may take up to 15 min.)
9. Configure the DNS resolution according to these instructions: [Setup DNS](dns.md)


## Deployment of the DSS services

During the successful run, the terraform job has created a new [workspace](https://github.com/interuss/dss/tree/master/build/workspace)
for the cluster. The new workspace name corresponds to the cluster context. The cluster context
can be retrieved by running `terraform output` in your infrastructure  folder (ie /deploy/infrastructure/personal/terraform-aws-dss-dev).

It contains scripts to operate the cluster and setup the services.

1. Go to the new workspace `/build/workspace/${cluster_context}`.
2. Run `./get-credentials.sh` to login to kubernetes. You can now access the cluster with `kubectl`.
3. If using CockroachDB:
    1. Generate the certificates using `./make-certs.sh`. Follow script instructions if you are not initializing the cluster.
    1. Deploy the certificates using `./apply-certs.sh`.
4. If using Yugabyte:
    1. Generate the certificates using `./dss-certs.sh init`
    1. If joining a cluster, check `dss-certs.sh`'s [help](../../operations/certificates-management.md) to add others CA in your pool and share your CA with others pools members.
    1. Deploy the certificates using `./dss-certs.sh apply`.
5. Go to the tanka workspace in `/deploy/services/tanka/workspace/${cluster_context}`.
6. Run `tk apply .` to deploy the services to kubernetes. (This may take up to 30 min)
7. Wait for services to initialize:
    - On AWS, load balancers and certificates are created by Kubernetes Operators. Therefore, it may take few minutes (~5min) to get the services up and running and generate the certificate. To track this progress, go to the following pages and check that:
        - On the [EKS page](https://eu-west-1.console.aws.amazon.com/eks/home), the status of the kubernetes cluster should be `Active`.
        - On the [EC2 page](https://eu-west-1.console.aws.amazon.com/ec2/home#LoadBalancers:), the load balancers (1 for the gateway, 1 per cockroach nodes) are in the state `Active`.
8. Verify that basic services are functioning by navigating to https://your-gateway-domain.com/healthy.


## Clean up

1. Note that the following operations can't be reverted and all data will be lost.
2. To delete all resources, run `tk delete .` in the workspace folder.
3. Make sure that all [load balancers](https://eu-west-1.console.aws.amazon.com/ec2/home#LoadBalancers:) and [target groups](https://eu-west-1.console.aws.amazon.com/ec2/home#TargetGroups:) have been deleted from the AWS region before next step.
4. `terraform destroy` in your infrastructure folder.
5. On the [EBS page](https://eu-west-1.console.aws.amazon.com/ec2/home#Volumes:), make sure to manually clean up the persistent storage. Note that the correct AWS region shall be selected.
