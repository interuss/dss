# AWS-1 CI deployment

This module deploys a DSS to a Kubernetes cluster in AWS. It is primarily by our [CI](../../../../.github/workflows/dss-deploy.yml).
See [test.sh](test.sh) for the complete list of actions.

## Terraform state

The terraform backend is configured to be shared using a S3 bucket. (see [`main.tf`](./main.tf)).

## Debugging

In case of issue, it is possible to connect to the cluster and retrieve the terraform state to manage it
locally.

### Connection to the cluster

To connect to the cluster, authenticate yourself to the AWS account. 
Run the following command to load the kubernetes config:
```
aws eks --region us-east-1 update-kubeconfig --name dss-ci-aws-ue1
```
Call the kubernetes cluster using `kubectl`

#### Add other roles

Access to the cluster is managed using the config map `aws-auth`. 
Its definition is managed in [`kubernetes_admin_access.tf`](./kubernetes_admin_access.tf).
Currently only the user who bootstrapped the cluster and the ones assuming 
the administrator role (see [`local_variables.tf`](./local_variables.tf)) have access.

### Run terraform locally

In case of failure, a user with administrator role can take over the deployment by cloning this 
repository and retrieving the current deployment state by running the following command:

```
terraform init
```

At this point, the user can replay or clean the deployment as if it was the CI runner.
