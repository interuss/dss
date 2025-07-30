# Deploy Kubernetes cluster

1. Create a new folder in `/deploy/infrastructure/personal/` named for instance `terraform-cloud-dss-dev`.
2. Copy main.tf, output.tf and variables.gen.tf to the new folder.
3. Copy `terraform.dev.example.tfvars` and rename to `terraform.tfvars`
4. Check that your new directory contains the following files:
    - main.tf
    - output.tf
    - terraform.tfvars
    - variables.gen.tf
5. Set the variables in `terraform.tfvars` according to your environment. See [TFVARS.gen.md](TFVARS.gen.md) for variables descriptions.
   TODO: differnt depending on provider
6. In the new directory (i.e. /deploy/infrastructure/personal/terraform-cloud-dss-dev), initialize terraform: `terraform init`.
7. Run `terraform plan` to check that the configuration is valid. It will display the resources which will be provisioned.
8. Run `terraform apply` to deploy the cluster. (This operation may take up to 15 min.)
