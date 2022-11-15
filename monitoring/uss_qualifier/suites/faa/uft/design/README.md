# UFT Message Signing tests for SCD using Uss_Qualifier

UTM Field Test (UFT) has a requirement for USSes to sign their message requests
and responses in the SCD flow using IATF certificates provided by FAA.

This test suite helps to test the message signing by extending [suites.astm.utm.f3548_21](../../../astm/utm/f3548_21.yaml)
with scenarios that trigger reporting of message signing in the interactions with mock_uss.

## Test Harness Setup

The test setup includes running the following components on your local machine, as also shown in the [diagram](./InterUss%20Test%20Harness%20With%20Message%20Signing.png)

1. Uss Qualifier - This is the test driver that injects the operations for the SCD flow tests in USSes.
2. Dummy Auth - This is a dummy auth server that provides tokens for uss-to-uss and uss-to-dss communication. Runs as a dockerized container at port 8085.
3. DSS - This is a local DSS running in dockerized container, and is available at port 8082.
4. Mock USS - This is a mock implementation of a USS running as a dockerized container at port 8074. The message signing reports are generated using mock_uss.
5. USS-under-Test - This is the USS that needs to be tested.
6. Uss Qualifier Interface. USSes need to develop an interface for their USS
to interface with the test harness. Uss_Qualifer will inject operations through this interface. The spec to
implement is - [Strategic Coordination Test Data Injection](https://github.com/interuss/automated_testing_interfaces/blob/fa3a5f544161c408f50255630a23b670c74a67d1/scd/v1/scd.yaml)

The main idea behind the tests is that mock_uss will consume and validate all the requests and responses from USS-under-test.

Note - As different USSes have different implementations, it could happen that you might need to create a subscription in the area of the flights.

## Steps to run the test

1. Set your UssQualifier Interface implementation url in the [configuration file ](../../../../configurations/dev/faa/uft/local_message_signing.yaml )to run the UFT message signing tests.
The property to set is `resources.resource_declarations.flight_planners.specification.flight_planners.participant_id`
2. Run DSS and dummy-oauth using the script [run_locally.sh](../../../../../../build/dev/run_locally.sh)
    ```bash
    ./run_locally.sh
    ```
3. Run mock_uss using the script [run_locally_scdsc.sh](../../../../../mock_uss/run_locally_scdsc.sh)
    ```bash
   ./run_locally_scdsc.sh
    ```
   Note - This script has an environment variable MESSAGE_SIGNING. It is set to true to turn message signing validation on in mock uss.
4. Run well-known server with public key for mock_uss. [run_well_known.sh](../../../../../messagesigning/run_well_known.sh)
    ```bash
   ./run_well_known.sh
   ```
5. Prepare your USS to run with
   1. dummy-oauth for getting tokens. Depending on your setup - localhost(http://localhost:8085/token) or dockerized (http://host.docker.internal:8085/token). The server has a GET and POST endpoint for getting tokens.
   2. DSS - Depending on your setup - localhost(http://localhost:8082/dss) or dockerized (http://host.docker.internal:8082/dss)
6. Run Uss Qualifier Interface for your USS.
7. Run USS Qualifier tests using script [run_locally.sh](../../../../../uss_qualifier/run_locally.sh) with config
    ```bash
   ./run_locally.sh configurations.dev.faa.uft.local_message_signing
   ```

## Results

The USS Qualifier tests run produces two reports.
1. SCD tests report is generated under [uss_qualifier](../../../../../uss_qualifier).
The report file name is - report.json
2. Message signing test report is generated under [mock_uss/report](../../../../../mock_uss/report)
The report file name is - report_mock_uss_scdsc_messagesigning.json.

There will be 3 runs of tests, with the above two reports for each run.

### Positive tests -
A valid set of private/public keys will be provided under [keys folder](../../../../../messagesigning/keys)
A USS should pass all the uss_qualifier tests in this suite.
As well as the message_signing report should have no issues reported. A USS will provide both of these reports.
The message_signing report includes interactions and issues between the mock_uss and the USS-under-test.
No issues indicate the USS-under-test message-signed all its requests and responses.

For analysis of the message signing reports, please refer to [Analyzing the message signing report in README.md](../../../../../messagesigning/README.md)

### Negative tests -
Replace the private/public keys with invalid key pair, the message signing by mock_uss will be invalid.
The USS-under-test will respond with 403 to all requests from mock_uss. The Uss Qualifier tests will not pass.
The USS can provide the two reports, the message signing report would show 403 in interactions with mock_uss.

### No message signing -
You can set the MESSAGE_SIGNING env variable to false to switch off message-signing in mock_uss.
The USSes are expected to let the requests with no message signing pass with 200. All the UssQualifier tests will pass.
You can provide both the reports. The message_signing reports will show 200 response from the USS-under-test.
