#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR" || { echo "Failed to chdir to $SCRIPT_DIR" >&2; exit 1; }

protoc --go_out=../../pkg/grpc --go_opt=paths=source_relative \
       --go-grpc_out=../../pkg/grpc --go-grpc_opt=paths=source_relative \
       consensus.proto

