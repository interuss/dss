# openapi-to-go-server

## Overview

This folder contains a tool to automatically generate Go code for the types and endpoints (interface & server) defined in one or more OpenAPI YAML files, as well as an example of a complete, executable webserver application generated almost entirely automatically.

The automatically-generated Go server uses only the built-in `net/http` and `encoding/json` libraries to implement the webserver (no third-party dependencies).

## Usage

The primary entrypoint to the generation tool is `generate.py`, and a complete generation environment can be produced with the `Dockerfile`.

The script `generate_example.sh` demonstrates the usage of this tool to generate a nearly-complete Go server from the ASTM SCD & RID APIs; run it from the working directory containing it.  See the [example](./example) folder for more information.

## openapi-to-go-server architecture

The generate.py entrypoint first parses all specified APIs into the forms recognized by openapi-to-go-server using the routines in apis.py.  The two primary components in APIs are data types, which are parsed with the tools in data_types.py, and operations (endpoints) which are parsed with the tools in operations.py (both incidentally using small utilities in formatting.py).  Once the APIs have been parsed into openapi-to-go-server's preferred representations, rendering.py then produces Go code to form an api library including all specified APIs, as well as an example entrypoint and dummy implementation.

## Generated architecture

### common.gen.go

The primary generated artifact is a Go-code api package.  That root package specifies some shared data structures and tools in common.gen.go, but the bulk of the content is located in each of the subpackages within api, one subpackage per API.

### types.gen.go

Within an API's package, the data types specified by the OpenAPI are rendered in types.gen.go in a form that can be automatically serialized and deserialized with JSON.

### interface.gen.go

An API's implementation is abstracted from the HTTP server with the interface defined in interface.gen.go; this file contains a Request object, Response object, and method in an interface for each operation defined in the API, as well as constants describing the security requirements prescribed by the API.  The Request object for a given operation contains all the relevant information provided by the client in a strictly-typed form.  The Response object contains a field for each kind of response defined by the API -- an implementation is expected to populate exactly one of these fields.

### server.gen.go

All boilerplate code for handling generic incoming HTTP requests using an instance of the implementation interface defined above (and an Authorizer that evaluates security requirements) is located in server.gen.go.  An API-specific APIRouter object is defined, and each operation defined in the API is added as a method.  Near the end of the file, a function is included that creates an APIRouter instance including regex-based routes to each method.  The APIRouter's Handle method nearly matches the handler method required by http.Server, but it returns a boolean indicating whether the request was handled.  This enables multiple APIRouters to be used in a single HTTP server using the shared MultiRouter.

### main.gen.go

All of the content in the api package (and its subpackages) is intended to be rendered directly into the main codebase and regenerated when the APIs change.  The specific implementation for each API is anticipated to be created once, manually, and then updated manually when the APIs change (because this is where all the custom business logic is located).  However, to demonstrate the generated api package and to provide a one-time starting point for a business logic implementation, openapi-to-go-server also has the capability of generating an entrypoint (invokable via `go run .`) and dummy implementation for each API when the --example_folder flag is specified.
