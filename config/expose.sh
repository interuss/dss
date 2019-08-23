#!/bin/bash

set -e

if [[ "$#" -ne 1 ]]; then
    echo "Usage: $0 NAMESPACE"
    exit 1
fi

set -x

for i in {0..2}; do
    kubectl expose pod "cockroachdb-$i" --namespace "$NAMESPACE" --type LoadBalancer --name "crdb-node-$i" --port=26257 --target-port=26257
done
