# Simple standalone local test/evaluation/development instance

_Note that all deployment strategies below require the use of [Docker](https://docs.docker.com/v17.12/install/)._

## Architecture

![Architecture diagram for running local processes](../../assets/generated/run_locally_architecture.png)

When the simple standalone deployment below is used, it will construct a DSS
sandbox environment consisting of:
* A DSS instance, consisting of:
  * A single CockroachDB node running in insecure mode via docker, communicating
    on port 26257 internally and exposing a web admin console on [port
    8080](http://localhost:8080) externally
  * [gRPC backend](../../cmds/grpc-backend) listening by default on port 8081
    internally, configured to
    * Connect to a CockroachDB node (implicitly port 26257)
    * Validate access tokens with [the auth2.pem public
      key](../test-certs/auth2.pem)
    * Expect access tokens to specify an `aud` of `localhost`
  * [HTTPS gateway](../../cmds/http-gateway) listening on port 8082 externally
    and directing requests translated into gRPC to port 8081 internally
* A [Dummy OAuth server](../../cmds/dummy-oauth) exposing an endpoint at
  http://localhost:8085/token externally to generate dummy JWT access tokens
  that validate against [the auth2.pem public key](../test-certs/auth2.pem)

## Prerequisites

* Install [Docker](https://docs.docker.com/v17.12/install/)
* Install [docker-compose](https://docs.docker.com/compose/install/)

## Run

Simply execute [`./run_locally.sh`](run_locally.sh).  This will build the required
Docker images if necessary and then construct the system described above.

When this system is active (log messages stop being generated), the following
endpoints will be available:

* Dummy OAuth Server: http://localhost:8085/token
* DSS HTTP Gateway Server: http://localhost:8082/healthy
* CockroachDB web UI: http://localhost:8080

In a different window, run [`./check_dss.sh`](check_dss.sh) to run a
demonstration RID query on the system.  The expected output is an empty list of
ISAs (no ISAs have been announced).

To perform more complicated actions manually, see
[the Postman collection](postman_collection.json) in this folder (use with
[Postman](https://www.postman.com/downloads/)).

To stop the system, just press ctrl-c or cmd-c.

## Advanced

[`run_locally.sh`](run_locally.sh) is a thin wrapper around a `docker-compose`
command and all the `docker-compose` verbs may be passed to `run_locally.sh`.
The default verb is `up`, but, e.g., the system can be removed entirely with
`run_locally.sh down`.  See all `docker-compose` verbs
[here](https://docs.docker.com/compose/reference/overview/).

Specifically, after changing code for one or more of the services, make sure to
execute `run_locally.sh build` to incorporate the new changes into the local
deployment images.  Make sure to restart the local system to reflect changes to
the local deployment images.

## Troubleshooting

If one or more of the necessary ports are not available, identify the process
using a port with `lsof -i tcp:8080`.
