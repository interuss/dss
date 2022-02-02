# http-gateway

## Introduction

This http-gateway executable is a translation layer that exposes a HTTP interfaces and fulfills them by with RPCs to core-service via gRPC.  It requires a connection to a core-service instance and exposes a few HTTP services: [ASTM remote ID](../../interfaces/uastech/standards/remoteid), [auxiliary](../../pkg/api/v1/auxpb/aux_service.proto), and [ASTM strategic coordination](../../interfaces/astm-utm/Protocol) (if specified).

## Usage

For production deployment of this executable, see [the deployment documentation](../../build/README.md).

For experimentation on a local machine, see [the standalone instance documentation](../../build/dev/standalone_instance.md).

To run this executable directly on a local machine using Go rather than a Docker container, run something similar to the command below from the repo root folder:

```bash
go run ./cmds/http-gateway \
  -core-service localhost:8081 \
  -addr :8082 \
  -trace-requests \
  -enable_scd
```

### Prerequisites

#### core-service

To run correctly, http-gateway must be able to access a core-service instance.  See [the core-service documentation](../core-service/README.md) for instructions on how to run an instance.
