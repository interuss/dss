# Deploy a DSS instance on Amazon Web Services via terraform

This terraform module creates a Kubernetes cluster in Amazon Web Services using the Elastic Kubernetes Service (EKS)
and generates the tanka files to deploy a DSS instance.

## Getting started

### Prerequisites

Download & install the following tools to your workstation:

1. Install [terraform](https://developer.hashicorp.com/terraform/downloads).
2. Install tools from [Prerequisites](index.md)
3. Install AWS specific tools:
   1. Install and initialize [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html#getting-started-install-instructions).
       1. Confirm successful installation with `aws --version`.
   2. If you don't have an account, sign-up: https://aws.amazon.com/free/
   3. Configure terraform to connect to AWS using your account.
      1. We recommend to create an AWS_PROFILE using for instance `aws configure --profile aws-interuss-dss`
         Before running `terraform` commands, run once in your shell: `export AWS_PROFILE=aws-interuss-dss`
         Other methods are described here: https://registry.terraform.io/providers/hashicorp/aws/latest/docs#authentication-and-configuration


### Deployment of the Kubernetes cluster

!!! tip "Paths in the documentation"
    In the documentation, we often refer to path starting from the root (prefixed with `/`). This is to indicate that the path is relative to the root of the repository.

1. Create a new folder in `/deploy/infrastructure/personal/` for the deployment named for example: `terraform-aws-dss-dev`
2. From `/deploy/infrastructure/modules/terraform-aws-dss`, copy `main.tf`, `output.tf`, `variables.gen.tf` and `terraform.dev.example.tfvars` to the infrastructure personal folder.
3. In the infrastructure personal folder (eg /deploy/infrastructure/personal/terraform-aws-dss-dev):
    1. Rename `terraform.dev.example.tfvars` to `terraform.tfvars`.
    2. Check that the directory contains the following files:
        1. main.tf
        2. output.tf
        3. terraform.tfvars
        4. variables.gen.tf
    3. Set the variables in `terraform.tfvars` according to your environment. See [TFVARS.gen.md](https://github.com/interuss/dss/blob/master/deploy/infrastructure/modules/terraform-aws-dss/TFVARS.gen.md) for variables descriptions.
    4. Initialize terraform: `terraform init`.
    5. Run `terraform plan` to check that the configuration is valid. It will display the resources which will be provisioned.
    6. Run `terraform apply` to deploy the cluster. (This operation may take up to 15 min.)
4. Configure the DNS resolution according to these instructions:

=== "Terraform managed"
    If your DNS zone is managed on the same account, it is possible to instruct terraform to create and manage it with the rest of the infrastructure.

    **For Elastic Kubernetes Service (AWS)**, create the zone in your aws account and set the `aws_route53_zone_id`
    variable with the zone id. Entries will be automatically created by terraform.
    Note that the domain or the sub-domain managed by the zone must be properly delegated by the parent domain.
    See instructions for [subdomains delegation](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/CreatingNewSubdomain.html#UpdateDNSParentDomain)

=== "Manual setup"
    If DNS entries are managed manually, set them up manually using the following steps:
    1. Retrieve IP addresses and expected hostnames: `terraform output`
        Example of expected output:
           ```
           crdb_addresses = [
               {
                   "address" = "34.65.15.23"
                   "expected_dns" = "0.interuss.example.com"
               },
               {
                   "address" = "34.65.146.56"
                   "expected_dns" = "1.interuss.example.com"
               },
               {
                   "address" = "34.65.191.145"
                   "expected_dns" = "2.interuss.example.com"
               },
           ]
           gateway_address = {
               "address" = "35.186.236.146"
               "expected_dns" = "dss.interuss.example.com"  
               "certificate_validation_dns" = [
                {
                  "managed_by_terraform" = false
                  "name" = "_6e246283563dcf58e7ed.interuss.example.com."
                  "records" = [
                     "_6e246283563dcf58e7ed.xxxxx.acm-validations.aws.",
                  ]
                  "type" = "CNAME"
                },
               ]
           }
           ```
    2. Create the following DNS A entries to point to the static ips indicated in the output:
      - `crdb_addresses[*].expected_dns`
      - `gateway_address.expected_dns`
    3. Create the entries for SSL certificate validation according to the information provided
      in `gateway_address.certificate_validation_dns`.

## Next steps

Proceed to [pooling configuration](../pooling/index.md).
