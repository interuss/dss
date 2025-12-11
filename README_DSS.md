
## DSS

Conceptual background on the DSS and services it supports may be found [here](./concepts.md).

## Simplified architecture

### Overview
![Simplified architecture diagram](./docs/assets/generated/simple_architecture.png)

A "DSS Region" consists of one or more DSS instances sharing the same
DSS Airspace Representation (DAR) by forming a single CockroachDB or Yugabyte
cluster. In the simplified diagram above, two DSS instances share the
same DAR via CRDB certificates and configuration which means the two
Core Services may be used interchangeably.  USS 1 chooses to use only
instance 1 while USS 2 uses both instances for improved resilience to
failures.

### Core Service

Component responsible for all the business logic as well as
authentication. This backend talks directly to CockroachDB or Yugabyte.
See the API specifications:
- [ASTM F3411-19 (Network Remote ID)](interfaces/rid/v1)
- [ASTM F3411-22a (Network Remote ID)](interfaces/rid/v2)
- [ASTM F3548-21 (Strategic Coordination)](interfaces/astm-utm)

### CockroachDB (CRDB)

Individual CockroachDB nodes hosting sharded data of the DAR. More information about CockroachDB
[here](https://www.cockroachlabs.com/docs/stable/architecture/overview.html).

The current implementation relied historicaly on CockroachDB for data storage and synchronization
between DSS participants, but is migrating to Yugabyte with both version supported. See [implementation details](implementation_details.md) for
more information.

### Yugabyte

Individual Yugabyte nodes hosting sharded data of the DAR. More information about Yugabyte
[here](https://www.yugabyte.com/).

The current implementation relied historically on CockroachDB for data storage and synchronization
between DSS participants but is migrating to Yugabyte with both version supported. See [implementation details](implementation_details.md) for
more information.

The DSS temporarily requires a patched version of Yugabyte, the reader may find [here](build/yugabyte_image_building.md) instructions on how to build it.


### DB Manager

Component responsible for performing orderly database schema migrations, including database bootstrapping. Invoked automatically on first run along with the CRDB cluster-init job, or when requested manually by a user.

## Directories of Interest:
*   [`build/`](build) has all of the configuration required to build and
    deploy a DSS instance. The README in that directory contains more
    information.
*   [`pkg/`](pkg) contains all of the source code for the DSS. See the
    README in that directory for more information.
*   [`cmds/`](cmds) contains entry points and docker files for the
    actual binary (`core-service`)
*   [`test/`](test) contains some tests; see [the README](test/README.md)
    for more information about testing.
