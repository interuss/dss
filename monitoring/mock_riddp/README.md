`mock_riddp` is a mock Remote ID Display Provider implementation.  It provides a
development-level web server that responds to requests to the USS endpoints
defined in the ASTM remote ID standard in a standards-compliant manner, and it
also includes an observation interface in accordance with the
[RID automated testing interface](../../interfaces/automated-testing/rid/README.md).

## Fully mocking an RID system

![Nominal RID system](../../assets/rid_fully_mocked.png)

An entire remote ID ecosystem (as described in the diagram above) can be deployed on a single local machine by following the instructions below.

1. Deploy DSS instance, including Dummy OAuth server: from `build/dev`, run `./run_locally.sh`
1. Deploy mock RID Service Provider
    1. Configure mock RID Service Provider to have Display Providers contact it at `host.docker.internal`: `export MOCK_RIDSP_TOKEN_AUDIENCE=host.docker.internal`
    1. From `monitoring/mock_ridsp`, run `./run_locally.sh`
1. Deploy mock RID Display Provider: from this folder, run `./run_locally.sh`
1. Run `rid_qualifier` configured to test this system: from `monitoring/rid_qualifier`, run `./test_fully_mocked_local_system.sh`
