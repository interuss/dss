# benchmarker release tests

This folder contains benchmarker configurations intended for characterizing performance of DSS releases in reference deployments in reference conditions.

Each `<test>_<pool>.jsonnet` configuration applies a [monitoring benchmarker](https://github.com/interuss/monitoring/tree/main/monitoring/benchmarker) test (`congested_area`, `separated_ops`, `typical_load`, `isas_uncontended`) to one pool of release DSS instances (`crdb` or `ybdb`, see [library](./library)).

## Prerequisites

The release clusters must be deployed and their services healthy (see [release/README.md](../../README.md)).

## Run

1. Spawn an AWS EC2 instance in `us-east-2`:
    * Instance type: `m5.2xlarge`
    * AMI: Debian 13

2. Connect to the instance with SSH and run everything as root:

    ```shell
    sudo -i
    ```

3. Clone the repository:

    ```shell
    apt-get update
    apt-get install -y git
    cd /opt
    git clone https://github.com/interuss/dss.git
    cd dss
    ```

4. Configure the images (both default to the values below):

    ```shell
    export MONITORING_IMAGE=docker.io/interuss/monitoring:latest
    export DUMMY_OAUTH_IMAGE=docker.io/interuss/dummy-oauth
    ```

5. Run the benchmarks:

    ```shell
    ./release/scripts/run-benchmarks.sh
    ```

    The script:
    * installs Docker if missing,
    * starts a dummy OAuth server,
    * runs the pools in parallel, and the configurations of a pool sequentially; a pool stops at its first failed configuration.

6. Retrieve the results, e.g. from your computer:

    ```shell
    scp -r <user>@<instance>:/opt/dss/release/tests/benchmarker/output .
    ```

7. Terminate the instance.

## Results

Each configuration writes to `output/<test>_<pool>/`:

* `congested_area`, `separated_ops`, `typical_load`:
    * `report.json`: raw report
    * `scalability_curve.png`: flight throughput, failures, create operational intent latency, DSS queries/s and USL fit, one subfigure per DSS instance
    * `timeline/index.html`: timeline of flights and operational intent operations, one `timeline/s<N>.html` per scenario
    * `usslogset.json`: queries performed
    * `intermediate<N>/`: `report.json`, `timeline/` and `scalability_curve.png` generated after each DSS instance but the last
* `isas_uncontended`:
    * `report.json`: raw report
    * `throughput_with_latency.png`: flight throughput, one series per DSS instance

Logs of each pool are written to `release/logs/<pool>.benchmarker.log`.
