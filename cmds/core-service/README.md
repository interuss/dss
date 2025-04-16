# core-service

## Introduction

This core-service executable is the main application logic of the DSS.  It requires a connection to a CockroachDB
database and exposes a few REST services: [ASTM remote ID](../../interfaces/rid),
[auxiliary](../../interfaces/aux_), and [ASTM strategic coordination](../../interfaces/astm-utm/Protocol) (if specified).

## Usage

For production deployment of this executable, see [the deployment documentation](../../build/README.md).

For experimentation on a local machine, see [the standalone instance documentation](../../build/dev/standalone_instance.md).

To run this executable directly on a local machine using Go rather than a Docker container, run something similar to the command below from the repo root folder:

```bash
go run ./cmds/core-service \
  -datastore_host localhost \
  -public_key_files build/test-certs/auth2.pem \
  -log_format console \
  -dump_requests \
  -accepted_jwt_audiences localhost \
  -enable_scd \
  -allow_http_base_urls
```

### Prerequisites

#### CockroachDB cluster

To run correctly, core-service must be able to [access](../../pkg/datastore/flags/flags.go) a CockroachDB cluster.  Provision of this cluster is handled automatically for a local development environment if following [the instructions for a standalone instance](../../build/dev/standalone_instance.md).  Or, a CockroachDB instance can be created manually with:

```bash
docker container run -p 26257:26257 -p 8080:8080 --rm cockroachdb/cockroach:v24.1.3 start-single-node --insecure
```

#### Database configuration

Once an initialized CockroachDB cluster is available, the necessary databases within the CRDB cluster must be created/configured properly.  This can be accomplished with [migrate_local_db.sh](../../build/dev/migrate_local_db.sh), as documented in the [standalone instance documentation](../../build/dev/standalone_instance.md), when using the standard standalone development DSS instance, or it can be accomplished manually with commands similar to those below starting from the repo root folder:

```bash
go run ./cmds/db-manager migrate \
  --schemas_dir ./build/db_schemas/rid \
  --db_version latest \
  --datastore_host localhost
go run ./cmds/db-manager migrate \
  --schemas_dir ./build/db_schemas/scd \
  --db_version latest \
  --datastore_host localhost
```
