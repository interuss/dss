#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR" || { echo "Failed to chdir to $SCRIPT_DIR" >&2; exit 1; }

mapfile -t PROTO_FILES < <(find . -type f -name '*.proto' -print)

if [ "${#PROTO_FILES[@]}" -eq 0 ]; then
  echo "No .proto files found under $SCRIPT_DIR" >&2
  exit 1
fi

echo "Found ${#PROTO_FILES[@]} .proto files. Generating..."

# Run protoc once for all files for efficiency; keep source_relative paths
if ! protoc --go_out=../../pkg/grpc --go_opt=paths=source_relative \
            --go-grpc_out=../../pkg/grpc --go-grpc_opt=paths=source_relative \
            "${PROTO_FILES[@]}"; then
  rc=$?
  echo "protoc failed with exit code ${rc}" >&2
  exit $rc
fi

echo "protoc generation completed successfully."
