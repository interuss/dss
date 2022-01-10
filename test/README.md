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
[`docker_e2e.sh`](docker_e2e.sh) script in this folder sets up a full
environment and runs a set of tests in that environment.  Docker is the only
prerequisite to running this end-to-end test on your local system.

### Running a subset of tests
To test a specific test in the [prober](../monitoring/prober) test suite,
simply add its name as the first argument to `docker_e2e.sh`.  For example:
```shell script
./docker_e2e.sh scd/test_constraint_simple.py
./docker_e2e.sh scd/test_constraint_simple.py::test_constraint_does_not_exist_get
```

### Examining Core Service logs
After a `docker_e2e.sh` run, the Core Service logs are automatically captured
to [core-service-for-testing.log](../core-service-for-testing.log).

## Lint checks
One of the continuous integration presubmit checks on this repository checks Go
style with a linter.  To run this check yourself, run the following command in
the root folder of this repo:
```shell script
make lint
```

## Interoperability tests
The [interoperability folder](../monitoring/interoperability) contains a test suite that
verifies interoperability between two DSS instances in the same region; see
[the README](../monitoring/interoperability/README.md) for more information.
