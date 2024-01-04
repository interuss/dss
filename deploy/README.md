# DSS Deployment

This folder contains the increments toward the new deployment approach as described in [#874](https://github.com/interuss/dss/issues/874).

The infrastructure folder contains the terraform modules to deploy the DSS to kubernetes clusters of various cloud providers:

- Amazon Web Services: [terraform-aws-dss](./infrastructure/modules/terraform-aws-dss/README.md)
- Google Cloud Engine: [terraform-google-dss](./infrastructure/modules/terraform-google-dss/README.md)

The service folder contains the scripts required to deploy the DSS to a Kubernetes cluster:

- Helm Charts: [services/helm-charts](./services/helm-charts)
- Tanka: [../build/deploy/](../build/deploy)
