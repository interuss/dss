# DSS testing

All of the tests below except the Interoperability tests are run as part of
continuous integration before a pull request is approved to be merged.

## Unit tests
Source code is often accompanied by `*_test.go` files which define unit tests
for the associated code.  All unit tests for the repo may be run with the
following command from the root folder of the repo:
```shell script
make test
```
The above command skips the CockroachDB tests because a `store-uri` argument is
 not provided.  To perform the CockroachDB tests, run the following command
 from the root folder of the repo:
```shell script
make test-cockroach
```

## Integration tests
For tests that benefit from being run in a fully-constructed environment, the
`make test-e2e` from the repo root folder sets up a full environment and runs
the prober tests in that environment.  Docker is the only  prerequisite to
running this end-to-end test on your local system.

For repeated tests without changes to the DSS, the local DSS instance can be
brought up initially with `make start-locally`, then the prober tests can be run
repeatedly with `make probe-locally` without needing to rerun
`make start-locally`.  To capture DSS logs, run `make collect-local-logs`.  To
bring down the local DSS instance at the conclusion of testing, run
`make stop-locally` or `make down-locally`.

When developing or troubleshooting features with the help of the [USS qualifier](https://github.com/interuss/monitoring/tree/main/monitoring/uss_qualifier)
where it might be useful to quickly iterate on changes both in the DSS and in a qualifier scenario,
an option is to:

1. build the DSS image locally via `make build-dss`
2. use the locally built version (`interuss-local/dss:latest`) in the [docker-compose file](https://github.com/interuss/monitoring/blob/843e69a166e6fb76459ebcda171dcd77a26ea5dc/build/dev/docker-compose.yaml#L46) that defines the qualifier's local test environment
3. start or restart the local USS qualifier deployment via `make restart-all` (in the [monitoring repository](https://github.com/interuss/monitoring/blob/843e69a166e6fb76459ebcda171dcd77a26ea5dc/Makefile#L116))
4. run the USS qualifier with a pre-packaged configuration such as `f3548_self_contained` or `dss_probing` in the monitoring repo. Eg, `./run_locally.sh configurations.dev.f3548_self_contained` from within the [monitoring/uss_qualifier](https://github.com/interuss/monitoring/tree/main/monitoring/uss_qualifier) directory of the qualifier repository.


### Running a subset of tests
To run a specific test in the [prober](../monitoring/prober) test suite,
simply add its name as the first argument to the script to run prober locally
(this is the same script `make probe-locally` uses).  For example:
```shell script
./build/dev/probe_locally.sh scd/test_constraint_simple.py
./build/dev/probe_locally.sh scd/test_constraint_simple.py::test_ensure_clean_workspace
```

### Examining Core Service logs
After a `make probe-locally` run, the Core Service logs can be examined in the
Core Service container (usually `dss_sandbox-local-dss-core-service-1`) or
dumped to core-service-for-testing.log using `make collect-local-logs`.

## Continuous integration
The other tests involved in continuous integration presubmit checks are
described in [the continuous integration folder](../.github/workflows/CI.md).
