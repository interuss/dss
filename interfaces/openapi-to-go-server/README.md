# openapi-to-go-server

## Overview

This folder contains a tool to automatically generate Go code for the types and endpoints (interface & server) defined in one or more OpenAPI YAML files, as well as an example of a complete, executable webserver application generated almost entirely automatically.

The automatically-generated Go server uses only the built-in `net/http` and `encoding/json` libraries to implement the webserver (no third-party dependencies).

## Usage

The primary entrypoint to the generation tool is `generate.py`, and a complete generation environment can be produced with the `Dockerfile`.

The script `generate_example.sh` demonstrates the usage of this tool to generate a nearly-complete Go server from the ASTM SCD & RID APIs; run it from the working directory containing it.  See the [example](./example) folder for more information.
