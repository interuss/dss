# DSS release candidate validation procedure

This procedure validates the image of a DSS release candidate. It deploys it in two DSS pools, one per datastore, each with one DSS instance on AWS and one on Google Cloud. It then runs the prober, the USS qualifier and the benchmarker against them, and packages the results.

| Cluster                   | Cloud                           | Datastore   | Deployed with | Nodes               |
| ------------------------- | ------------------------------- | ----------- | ------------- | ------------------- |
| `release-aws-dss-crdb`    | AWS (`us-east-2`)               | CockroachDB | Tanka         | 3 × `m5.xlarge`     |
| `release-google-dss-crdb` | Google Cloud (`us-central-1-a`) | CockroachDB | Helm          | 3 × `n2-standard-4` |
| `release-aws-dss-ybdb`    | AWS (`us-east-2`)               | Yugabyte    | Tanka         | 3 × `m5.xlarge`     |
| `release-google-dss-ybdb` | Google Cloud (`us-central-1-a`) | Yugabyte    | Helm          | 3 × `n2-standard-4` |

The clusters are defined in [infrastructure/](infrastructure), and the tests in [tests/](tests).

## Prerequisite

### 1. Tools

terraform, aws, gcloud, kubectl, helm, tk (Tanka) and envsubst (package gettext-base).

### 2. Cloud access

```shell
aws login
aws configure --profile aws-interuss-dss

gcloud auth login
cat ~/.config/gcloud/configurations/config_default   # check that the project is the validation project
unset GOOGLE_CLOUD_PROJECT                            # avoids auth issues if other projects are configured
```

### 3. Configure environment variables

_Note: change `v0.22.0` below to the DSS release candidate version being validated._

```shell
export IMAGE=docker.io/interuss/dss:v0.22.0
export ZONE_ID=(AWS zone id)
export GOOGLE_PROJECT_NAME=(google project name)
export ZONE_NAME=(google zone name)
export AWS_PROFILE=(if needed)
```

Run the commands below from the root of the repository. Each step writes its logs to release/logs/<cluster>.<step>.log.

### 4. Test machine

Steps 4 to 6 of the procedure run on an AWS EC2 instance:

1. Spawn an AWS EC2 instance in `us-east-2`, with instance type `m5.2xlarge` and AMI Debian 13.

2. Connect to the instance with SSH, then run everything as root:

    ```shell
    sudo -i
    apt-get update
    apt-get install -y git curl zip unzip
    cd /opt
    git clone https://github.com/interuss/dss.git
    cd dss
    ```

3. Configure environment variables:

    _Note: change `v0.22.0` below to the DSS release candidate version being validated._

    ```shell
    export IMAGE=docker.io/interuss/dss:v0.22.0
    export MONITORING_IMAGE=docker.io/interuss/monitoring:latest
    export DUMMY_OAUTH_IMAGE=docker.io/interuss/dummy-oauth
    ```

Docker is installed by the scripts if missing.

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

This deploys the DSS instances, then waits until https://<DSS>/healthy returns ok on the four instances. While it waits, it lists the pods that are not ready.

Done when it prints `All 4 clusters healthy`.

### 4. Run the functional tests

On the test machine:

```shell
./release/scripts/run-tests.sh
```

This starts a local dummy OAuth server on port 8085. It then runs the prober against each instance, and the USS qualifier configurations crdb and ybdb against each pool.

Done when it prints `ALL PASSED`. If it prints `FAILED`, the image did not pass validation. Keep the logs, and continue to step 6 to collect the results.

### 5. Benchmark performance

On the test machine:

```shell
./release/scripts/run-benchmarks.sh
```

This starts a local dummy OAuth server on port 8085. It then runs each [benchmarker](https://github.com/interuss/monitoring/tree/main/monitoring/benchmarker) configuration `tests/benchmarker/<test>_<pool>.jsonnet`, which applies a test (`congested_area`, `separated_ops`, `typical_load`, `isas_uncontended`) to one pool (`crdb` or `ybdb`, see [tests/benchmarker/library](tests/benchmarker/library)). Pools are benchmarked in parallel, and the configurations of a pool sequentially; a pool stops at its first failed configuration. Do not run it at the same time as step 4.

Done when it prints `ALL DONE`. Logs of each pool are written to release/logs/<pool>.benchmarker.log, and each configuration writes to `tests/benchmarker/output/<test>_<pool>/`:

* `congested_area`, `separated_ops`, `typical_load`:
    * `report.json`: raw report
    * `scalability_curve.png`: flight throughput, failures, create operational intent latency, DSS queries/s and USL fit, one subfigure per DSS instance
    * `timeline/index.html`: timeline of flights and operational intent operations, one `timeline/s<N>.html` per scenario
    * `usslogset.json`: queries performed
    * `intermediate<N>/`: `report.json`, `timeline/` and `scalability_curve.png` generated after each DSS instance but the last
* `isas_uncontended`:
    * `report.json`: raw report
    * `throughput_with_latency.png`: flight throughput, one series per DSS instance

### 6. Compile the results

On the test machine:

```shell
./release/scripts/compile-results.sh
```

This packages the prober, USS qualifier and benchmarker results.

Done when `release/output/<version>.zip` exists. Retrieve it, e.g. from your computer:

```shell
scp <user>@<instance>:/opt/dss/release/output/<version>.zip .
```

Then terminate the test machine.

### 7. Destroy the clusters

```shell
./release/scripts/destroy-clusters.sh
```

This uninstalls the services, deletes the persistent volumes, then runs terraform destroy on each cluster.

Done when it prints `All 4 clusters destroyed` with no warning. If there was a warning, check <cluster>.uninstall.log and <cluster>.destroy.log. Finally, check the AWS and Google Cloud consoles for leftover resources, especially load balancers and disks.
