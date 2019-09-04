#!/bin/bash

set -e

if [[ "$#" -ne 1 ]]; then
    echo "Usage: $0 NAMESPACE"
    exit 1
fi

NAMESPACE="$1"

set -x

# Paths to directories in which to store certificates and generated YAML files.
DIR="$(pwd)"
CLIENTS_CERTS_DIR="$DIR/generated/$NAMESPACE/client_certs_dir"
NODE_CERTS_DIR="$DIR/generated/$NAMESPACE/node_certs_dir"
JWT_PUBLIC_CERTS_DIR="$DIR/jwt-public-certs"

# Delete previous secrets in case they have changed.
kubectl create namespace "$NAMESPACE" || true

kubectl delete secret cockroachdb.client.root || true
kubectl delete secret cockroachdb.client.root --namespace "$NAMESPACE" || true
kubectl delete secret cockroachdb.node --namespace "$NAMESPACE" || true
kubectl delete secret dss.public.certs --namespace "$NAMESPACE" || true

kubectl create secret generic cockroachdb.client.root --from-file "$CLIENTS_CERTS_DIR"
kubectl create secret generic cockroachdb.client.root --namespace "$NAMESPACE" --from-file "$CLIENTS_CERTS_DIR"
kubectl create secret generic cockroachdb.node --namespace "$NAMESPACE" --from-file "$NODE_CERTS_DIR"
kubectl create secret generic dss.public.certs --namespace "$NAMESPACE" --from-file "$JWT_PUBLIC_CERTS_DIR"
