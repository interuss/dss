# CI infrastructure

This directory is a copy of [`build/dev` from interuss/monitoring](https://github.com/interuss/monitoring/tree/main/build/dev), which deploys a local pool of DSS instances behind a load balancer, with a dummy OAuth server.

It is a temporary solution, used by CI to run the tests against a multi-node raft cluster, while we define the testing stack for the raftstore. Until then, we aim to keep this aligned with the monitoring tooling.

The Docker Compose projects are named `ci_infra_*` instead of `local_infra_*` so that this deployment does not interfere with the monitoring one or with the standalone DSS instance when stopped. They share networks and published ports though, so only one of them can be running at a time.

The load balancer is also exposed on port `8082` as `core-service`, like the standalone DSS instance, so that `make probe-locally`, `qualify-locally` and `security-locally` can run against it. See the `raft-cluster` entry of the `dss-tests-matrix` job in [ci.yml](../../../.github/workflows/ci.yml).
