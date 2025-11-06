# Infrastructure

Infrastructure provides instructions and tooling to easily provision a Kubernetes cluster and cloud resources (load balancers, storage...) to a cloud provider. The resulting infrastructure meets the [Pooling requirements](../architecture.md#objective). Terraform modules are provided for:

* [Amazon Web Services (EKS)](terraform-aws-dss)
* [Google (GKE)](terraform-google-dss)
