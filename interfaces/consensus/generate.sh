#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

protoc --go_out=../../pkg/grpc --go_opt=paths=source_relative \
       --go-grpc_out=../../pkg/grpc --go-grpc_opt=paths=source_relative \
       consensus.proto

