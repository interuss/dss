# USS to USS Communication and Synchronization [![Build Status](https://dev.azure.com/astm/dss/_apis/build/status/interuss.dss?branchName=master)](https://dev.azure.com/astm/dss/_build/latest?definitionId=2&branchName=master) [![GoDoc](https://godoc.org/github.com/interuss/dss?status.svg)](https://godoc.org/github.com/interuss/dss)
This repository contains a simple and open service used by separate UAS
Service Suppliers (USSs), often in different organizations, to
communicate information about UAS operations and coordinate with each
other.  This service is a Discovery and Synchronization Service (DSS) as
described in the ASTM remote ID standard.  This flexible and distributed
system is used to connect multiple USSs operating in the same general
area to share information while protecting operator and consumer
privacy. The system is focused on facilitating communication amongst
actively operating USSs without details about UAS operations stored or
processed in the DSS.

## Simplified architecture

### Overview
![Simplified architecture diagram](assets/generated/simple_architecture.png)

A "DSS Region" consists of one or more DSS instances sharing the same
DSS Airspace Representation (DAR) by forming a single CockroachDB
cluster.  In the simplified diagram above, two DSS instances share the
same DAR via CRDB certificates and configuration which means the two
HTTPS frontends may be used interchangeably.  USS 1 chooses to use only
instance 1 while USS 2 uses both instances for improved resilience to
failures.

### HTTPS frontend

Serves as an HTTPS gateway to the business logic, translating between
HTTPS request and gRPC to allow users to communicate to the DSS via
simple HTTPS calls. This code is currently generated via
[grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway), and does
not do much other than translation.  See the [API specification
here](https://tiny.cc/dssapi_rid).

### gRPC backend

Component responsible for all the business logic as well as
authentication. This backend talks directly to CockroachDB.

### CockroachDB (CRDB)

Individual CockroachDB nodes hosting sharded data of the DAR. More information about CockroachDB
[here](https://www.cockroachlabs.com/docs/stable/architecture/overview.html).

## Directories of Interest:
*   [`build/`](build) has all of the configuration required to build and
    deploy a DSS instance. The README in that directory contains more
    information.
*   [`pkg/`](pkg) contains all of the source code for the DSS. See the
    README in that directory for more information.
*   [`cmds/`](cmds) contains entry points and docker files for the
    actual binaries (the `http-gateway` and `grpc-backend`)

## Notes

*   Currently this branch only supports remote ID APIs and
    functionality.
*   The current implementation relies on CockroachDB for data storage
    and synchronization between DSS participants.  See [implementation
    details](implementation_details.md) for more information.
