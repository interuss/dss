# [Continuous integration](ci.yml)

Before a pull request can be merged into the master branch, it must pass all automated tests for the repository.  This document describes the tests and how to run them locally.

## Repository hygiene (`make check-hygiene`)

### Python lint (`make python-lint`)

### Automated hygiene verification (`make hygiene`)

### uss_qualifier documentation validation (`make validate-uss-qualifier-docs`)

### Shell lint (`make shell-lint`)

### Go lint (`make go-lint`)

## DSS tests (`make check-dss`)

### Deployment infrastructure tests (`make evaluate-tanka`)

### Go unit tests (`make test-go-units`)

### Go unit tests with CockroachDB (`make test-go-units-crdb`)

### Build `dss` image (`make build-dss`)

### Build `monitoring` image (`make build-monitoring`)

### End-to-end test (`make test-e2e`)

Steps:

* `make start-locally` (build/dev/run_locally.sh)
* Run pytest in monitoring/prober (in `monitoring` container)

## `monitoring` tests (`make check-monitoring`)

### monitorlib tests (`make test` in monitoring/monitorlib)

### mock_uss tests (`make test` in monitoring/mock_uss)

Steps:

* Bring up geoawareness mock_uss
* Run geoawareness pytest

### uss_qualifier tests (`make test` in monitoring/uss_qualifier)

Steps:

* test_docker_fully_mocked.sh
