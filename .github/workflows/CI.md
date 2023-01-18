# [Continuous integration](ci.yml)

Before a pull request can be merged into the master branch, it must pass all automated tests for the repository.  This document describes the tests and how to run them locally.

## Repository hygiene (`make check-hygiene`)

### Python lint (`make python-lint`)

### Automated hygiene verification (`make hygiene`)

### Shell lint (`make shell-lint`)

### Go lint (`make go-lint`)

## DSS tests (`make check-dss`)

### Deployment infrastructure tests (`make evaluate-tanka`)

### Go unit tests (`make test-go-units`)

### Go unit tests with CockroachDB (`make test-go-units-crdb`)

### Build `dss` image (`make build-dss`)

### Tear down any pre-existing local DSS instance (`make down-locally`)

### Start local DSS instance (`make start-locally`)

### Probe local DSS instance (`make probe-locally`)

### Bring down local DSS instance (`make down-locally`)
