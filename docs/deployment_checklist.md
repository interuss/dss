# Deployment Checklist

This checklist outlines the major decisions and steps required to deploy a non-local (e.g. production, qualification) DSS instance.

## Preparation

* [ ] Review the [architecture requirements](architecture/index.md).
* [ ] Decide on the datastore you will use (CockroachDB or YugabyteDB). **All participants in a DSS Pool must use the same datastore**, so plan accordingly.
* [ ] Decide how and where you will deploy your DSS instances:
    * This repository provides Terraform configurations for [Amazon Web Services (EKS)](infrastructure/aws.md) and [Google Cloud (GKE)](infrastructure/google.md) to deploy a Kubernetes cluster (the infrastructure into which the Services will be deployed).
    * This repository provides [Tanka](https://github.com/interuss/dss/blob/master/deploy/services/tanka/) files and [Helm Charts](https://github.com/interuss/dss/blob/master/deploy/services/helm-charts/dss) to be used to deploy Services into a Kubernetes cluster. Terraform will automatically generate these configurations if needed.
    * You may also choose to deploy manually or use custom configuration tools.
    * Latency is an important factor to achieve good performances. Review the [latency documentation](architecture/latency.md) and plan with others participants of your DSS pool.
* [ ] Prepare sufficient resources for the services.
    * In particular, review the [CockroachDB recommendations](https://www.cockroachlabs.com/docs/v24.1/recommended-production-settings#cloud-specific-recommendations) and [YugabyteDB recommendations](https://docs.yugabyte.com/stable/deploy/checklist/#public-clouds); the datastore will consume the majority of the resources.
    * Example sizing is also describled in [sizing](architecture/sizing.md).


## Deployment

* [ ] Deploy the DSS instance by following the guides based on your previous infrastructure choices.
    * [ ] If needed, you will pool your DSS instance. Guides are available for [CockroachDB](operations/pooling-crdb.md) and [YugabyteDB](operations/pooling.md).
* [ ] If needed, [monitor metrics](operations/monitoring.md) of your DSS instance.
* [ ] If needed, track the availability of your DSS instance using [health checks](operations/healthchecks.md).
* [ ] Review the [database cleanup documentation](operations/cleanup.md) and enable cleanup cron jobs if required.
