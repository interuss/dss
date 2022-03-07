# Strategic Coordination and UAS Flight Authorisation

## Local qualification
1. Copy `config_run.json.dist` to `config_run.json`.
2. Edit the `injection_base_url` fields to the required target urls of the service providers to test.
3. Setup the dependencies:
    1. Start the local DSS instance and the Dummy OAuth server using [/build/dev/run_locally.sh](/build/dev/run_locally.sh)
4. Ensure that the service provider to be tested implements the [injection interfaces](/interfaces/automated-testing/scd).
5. Configure the service provider to connect to the local DSS instance and the Dummy OAuth server:
   1. Local DSS server url: `http://host.docker.internal:8082`
   2. Dummy OAuth server url: `http://host.docker.internal:8085/token`
   3. Dummy OAuth server public key: [`/build/test-certs/auth2.pem`](/build/test-certs/auth2.pem)
   4. Audience should correspond to the base url of the service provider to be tested.
6. Run the automated test: `bin/run.sh config_run.json`
7. Consult the report: [`report_scd.json`](../report_scd.json)
