# benchmarker release tests

This folder contains benchmarker configurations intended for characterizing performance of DSS releases in reference deployments in reference conditions.

To run a configuration, first host a dummy OAuth server locally on port 8085.  From the root of the `monitoring` repo:

```shell
make start-locally
```

Then, run benchmarker:

```shell
docker container run \
  --rm \
  -u "$(id -u):$(id -g)" \
  --network interop_ecosystem_network \
  -v "$(pwd):/benchmarker" \
  -w /app \
  interuss/monitoring:v0.33.0 \
  uv run python monitoring/benchmarker/benchmark.py \
    --config file:///benchmarker/congested_area.jsonnet \
    --output /benchmarker/output
```

Manually edit congested_area.jsonnet to change the `db_type` value to `ybdb` and rerun for YugabyteDB performance characterization.
