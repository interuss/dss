# openapi-to-go-server example

## Overview

This folder contains the output from the openapi-to-go-server tool (*.gen.go) as well as a few helpers to run the webserver via Docker.  After running [generate_example.sh](../generate_example.sh), run [run_example.sh](run_example.sh) to run the generated webserver.  Endpoints can be accessed at http://localhost:8080/scd as the base URL for the API defined at https://tiny.cc/dssapi_utm; for instance, navigate to http://localhost:8080/scd/dss/v1/operational_intent_references/foo to see a default JSON result.
