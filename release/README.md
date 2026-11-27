# DSS release candidate validation procedure

This procedure validates the image of a DSS release candidate. It deploys it in two DSS pools, one per datastore, each with one DSS instance on AWS and one on Google Cloud. It then runs the prober, the USS qualifier and the benchmarker against them, and packages the results.

This folder contains script helpers, deployment configurations, and test configurations to be run on DSS images for each release.
Scripts are not expected to be robust; in case of an error, manual intervention will be needed. They are only intended to assist someone with sufficient technical knowledge.

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

1. Spawn an AWS EC2 instance in `us-east-2`, with instance type `m5.2xlarge`, 200Gb of root disk and AMI Debian 13.

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

This runs terraform apply on the four clusters in parallel and configurations are stored in `personal/`. Existing configs in `personal/` are not removed: the script asks whether to proceed, then whether to overwrite their module and tfvars (terraform state is kept).

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

Done when it prints `ALL DONE`. Logs of each pool are written to release/logs/<pool>.benchmarker.log, and each configuration writes to `tests/benchmarker/output/<test>_<pool>/`.

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

This uninstalls the services, deletes the persistent volumes, then runs terraform destroy on each cluster. The `personal/` config of each successfully destroyed cluster is removed.

Done when it prints `All 4 clusters destroyed` with no warning. If there was a warning, check <cluster>.uninstall.log and <cluster>.destroy.log. Finally, check the AWS and Google Cloud consoles for leftover resources, especially load balancers and disks.

## Compatibility matrix

```shell
./release/scripts/run-compat-matrix.sh
```

This tests which upgrade paths are possible between released DSS versions, so that the compatibility matrix of the documentation can be updated. It runs locally and does not need the clusters.

The list of versions to test is the `VERSIONS` array at the top of the script, ordered from oldest to newest. The version being released must be added to it.

Please also ensure MONITORING_IMAGE is set to the current version used for releases and that the qualifier's configuration is up to date.

Any local DSS instance must be stopped first with `make stop-locally`, otherwise it interferes with the stack started by the script.

For each pair of versions A (older) to B (newer), the script starts a local stack with a single datastore shared by two `core-service` instances, one running version A and the other version B. The datastore is migrated to the latest schema of version B. Pairs where B is older than A are not tested and reported as not evaluated.

Each pair is then validated by running the prober against both instances and the USS qualifier against the pool, from `MONITORING_IMAGE`. A pair passes only if the three runs pass.

A local `dummy-oauth` image is built if it is not already present. The docker compose stack and the qualifier configuration are in [compat/](compat).

Done when it prints a summary matrix in the terminal, followed by a Markdown compatibility matrix to copy into [docs/operations/upgrades.md](../docs/operations/upgrades.md) and the GitHub release notes.

Container logs are available in `release/logs` and qualifier reports in `release/uss_qualifier_output/compat`.
