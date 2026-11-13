# DSS release validation

This procedure validates a DSS image. It deploys it in two DSS pools, one per datastore, each with one DSS instance on AWS and one on Google Cloud. It then runs the prober, the USS qualifier and the benchmarker against them, and packages the results.

| Cluster                   | Cloud                           | Datastore   | Deployed with | Nodes               |
| ------------------------- | ------------------------------- | ----------- | ------------- | ------------------- |
| `release-aws-dss-crdb`    | AWS (`us-east-2`)               | CockroachDB | Tanka         | 3 × `m5.xlarge`     |
| `release-google-dss-crdb` | Google Cloud (`us-central-1-a`) | CockroachDB | Helm          | 3 × `n2-standard-4` |
| `release-aws-dss-ybdb`    | AWS (`us-east-2`)               | Yugabyte    | Tanka         | 3 × `m5.xlarge`     |
| `release-google-dss-ybdb` | Google Cloud (`us-central-1-a`) | Yugabyte    | Helm          | 3 × `n2-standard-4` |

The clusters are defined in [infrastructure/](infrastructure), and the tests in [tests/](tests).

## Prerequisite

### 1. Tools

terraform, aws, gcloud, kubectl, helm, tk (Tanka), docker, envsubst (package gettext-base), curl and zip. Docker must be running, and port 8085 must be free.

### 2. Cloud access

```shell
aws login
aws configure --profile aws-interuss-dss

gcloud auth login
cat ~/.config/gcloud/configurations/config_default   # check that the project is the validation project
unset GOOGLE_CLOUD_PROJECT                            # avoids auth issues if other projects are configured
```

### 3. Configure environment variables

```shell
export IMAGE=docker.io/interuss/dss:v0.22.0
export ZONE_ID=(AWS zone id)
export GOOGLE_PROJECT_NAME=(google project name)
export ZONE_NAME=(google zone name)
export AWS_PROFILE=(if needed)
```

Run the commands below from the root of the repository. Each step writes its logs to release/logs/<cluster>.<step>.log.

## Procedure

### 1. Spawn the clusters (infrastructure layer)

```shell
./release/scripts/spawn-clusters.sh
```

This runs terraform apply on the four clusters in parallel and configurations are stored in `personal/`.

Done when it prints `All 4 clusters up`. If some issues are encountered during the deployment, it is possible to perform a `terraform apply` in the `personal/*` folders.

### 2. Configure clusters

```shell
./release/scripts/configure-clusters.sh
```

This fetches the Kubernetes credentials, generates new certificates, and makes the two instances of each pool trust each other.

Done when it prints `All 4 clusters prepared`.

After this step, services are ready to be deployed.

### 3. Deploy the services

```shell
./release/scripts/deploy-services.sh
```

This deploys the DSS, then waits until https://<DSS>/healthy returns ok on the four instances. While it waits, it lists the pods that are not ready.

Done when it prints `All 4 clusters healthy`.

### 4. Run the functional tests

```shell
./release/scripts/run-tests.sh
```

This starts a local dummy OAuth server on port 8085. It then runs the prober against each instance, and the USS qualifier configurations crdb and ybdb against each pool.

Done when it prints `ALL PASSED`. If it prints `FAILED`, the image did not pass validation. Keep the logs, and continue to step 6 to collect the results.

### 5. Benchmark performance

While the clusters are still up, run the benchmarker with the configurations in tests/benchmarker/ twice: once with db_type = 'crdb' and once with db_type = 'ybdb'.

Done when the outputs contain the four scenarios aws crdb, google crdb, aws ybdb and google ybdb.

### 6. Compile the results

```shell
./release/scripts/compile-results.sh
```

Done when `release/output/<version>.zip` exists, with this structure:

```
<version>/
├── cockroachdb/
│   ├── README.md
│   ├── prober/
│   │   ├── dss1-aws.txt
│   │   └── dss2-google.txt
│   └── uss_qualifier/
│       └── pooled_dss_probing/
└── yugabyte/
    └── (same structure)
```

The benchmarker outputs from step 5 are not included in the zip.

### 7. Destroy the clusters

```shell
./release/scripts/destroy-clusters.sh
```

This uninstalls the services, deletes the persistent volumes, then runs terraform destroy on each cluster.

Done when it prints `All 4 clusters destroyed` with no warning. If there was a warning, check <cluster>.uninstall.log and <cluster>.destroy.log, then check the AWS and Google Cloud consoles for leftover resources, especially load balancers and disks.
