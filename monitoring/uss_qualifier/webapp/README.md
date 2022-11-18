# USS Qualifier WebApp

# NOTE: WebApp has not yet been upgraded to support newer uss_qualifier architecture

`uss_qualifier webapp` is a development-level web server that provides endpoint to initialize the automated testing of the USS qualifier.
The local environment for the web server is setup by [run_locally.sh](run_locally.sh) script.

## Architecture

This project will create the following environment:

-   redis Container: A redis server to handle background tasks.

-   rq-worker Container: A container to run redis queue worker processes.

-   uss-host Container: A flask application which accepts KML files or flight states' json files as input, collect auth and config spec from the user and allow user to execute USS Qualifier test. Once task finishes successfully, report file becomes available to download from the UI.

-   mock-riddp, mock-rdisp and mock-scdsc containers: These are the USS mock containers brought up to handle RID and SCD specific user configurations' test runs.

This application accepts Auth2 credentials for user-specific login. The process to authenticate using Auth2 is described [here](LOGIN.md). If Auth2 credentials are not provided application uses `Local User` session by default.

The application provides [API endpoints](../../../interfaces/uss_qualifier/uss_qualifier.yaml) which are consumable by other applications. The thirdparty application should be configured with OAuth2 and client credentials. The process to enable OAuth2 for Postman is described [here](OAUTH2_CONFIG_POSTMAN.md)

### Input Files

Accepted input files are in the FullFlightRecord format.

When KML files are provided as input, `Flight records json` files are generated from the KML and kept into user specific `flight_records` folder on the RID host. Generated files are then available on the UI for selection during test execution.

### Test Execution

When the user requests the execution of a test, the `uss-host` sends this request as a message to the `redis` server Message Queue, which is then handled by the worker process running on the `rq-worker`. This job takes some amount of time, during this time UI keeps polling `uss-host` server to check for the job status. In the background, uss-host then checks the `Redis` container for the job progress and returns current status to the UI. Once job finishes successfully, uss-host generates resulting report from the job response and keep it in user specific `tests` folder on uss-host, and sends it as a response to the UI.

All the tests run by a user are then available to download from the UI.

## Run via docker-compose

Change the root directory to dss/.

Run [build/dev/run_locally.sh](../../../build/dev/run_locally.sh) to bring up local development DSS
instance.

Run the following command to start USS Qualifier webapp containers:

```bash
 docker-compose -f monitoring/uss_qualifier/webapp/docker-compose-webapp.yaml up
```

All the docker sub-commands such as `build`, `up`, `down`, `rm` etc. are also applicable to docker-compose and hence applied to all the containers listed in docker-compose file. Similary `--no-cache` flag can be used to prevent caching while rebuilding the image.

## Task Description

You can now test the application by uploading flight states json files. There must be at least one such input json file w.r.t each USS target provided in [Configuration object below](#user-config-information).
Input files are **required** to execute the task. Follow the instructions below to generate [Flight states input files](#flight-states-input-files).

## Flight states input files

RID host requires flight state input files to run the test executor. These files can be generated using one of the simulators for this type of resource.

## User Config information

uss_qualifier host needs a configuration object and an auth spec (single string e.g. `NoAuth()`) in order to run. Configuration object is a simple nested json string which should have at least a list of `injection_targets` and `observers`.

A task may take few minutes to finish. A new task can not be launched until current task ends.

Once the task is successful, resulting report can be retrieved from the web interface by clicking the `Get Result` button. A sample response report.json can be downloaded by ticking `Sample Report` checkbox.
