# Dummy OAuth

## Contents

This folder contains a development utility that generates OAuth access tokens compatible with ASTM APIs with the specified fields (and no security).

## Usage

The API for Dummy OAuth may be found [here](../../interfaces/dummy-oauth)

Dummy OAuth can be run directly on a development system with Go installed by starting in the repo root folder and:

```bash
go run ./cmds/dummy-oauth
```

To use the [Docker image](Dockerfile) for Dummy OAuth, leverage [build/dev/run_locally.sh](../../build/dev/run_locally.sh) starting from the root folder of the repo:

```bash
build/dev/run_locally.sh build local-dss-dummy-oauth
build/dev/run_locally.sh up -d local-dss-dummy-oauth
```

Get a token using an approach similar to this:

```bash
curl "http://localhost:8085/token?sub=uss1&intended_audience=uss2&scope=dss.read.identification_service_areas&issuer=dummy_oauth"
```

Token contents can be verified at https://dinochiesa.github.io/jwt/, and the signature can be validated with the [auth2.pem public key](../../build/test-certs/auth2.pem) by default.

Take down the Dummy OAuth instance like this:

```bash
build/dev/run_locally.sh down
```

## Development

The Dummy OAuth API scaffolding is generated automatically by [openapi-to-go-server](../../interfaces/openapi-to-go-server) using [the Dummy OAuth API](../../interfaces/dummy-oauth/dummy-oauth.yaml) via the command [`make apis`](../../Makefile) starting in the repo root folder.
