#!/bin/bash

set -e
set -x

# Paths to directories in which to store certificates and generated YAML files.
DIR="$(pwd)"
NAMESPACE="dss-main"
CLIENTS_CERTS_DIR="$DIR/generated/$NAMESPACE/client_certs_dir"
NODE_CERTS_DIR="$DIR/generated/$NAMESPACE/node_certs_dir"
CA_KEY_DIR="$DIR/generated/$NAMESPACE/ca_key_dir"
CA_CRT_DIR="$DIR/generated/$NAMESPACE/ca_crt_dir"
JWT_PUBLIC_CERTS_DIR="$DIR/jwt-public-certs"
UPLOAD_CA_KEY=false

# Delete previous secrets in case they have changed.
kubectl create namespace "$NAMESPACE" || true

kubectl delete secret cockroachdb.client.root || true
kubectl delete secret cockroachdb.client.root --namespace "$NAMESPACE" || true
kubectl delete secret cockroachdb.node --namespace "$NAMESPACE" || true
kubectl delete secret dss.public.certs --namespace "$NAMESPACE" || true

kubectl create secret generic cockroachdb.client.root --from-file "$CLIENTS_CERTS_DIR"
kubectl create secret generic cockroachdb.client.root --namespace "$NAMESPACE" --from-file "$CLIENTS_CERTS_DIR"
kubectl create secret generic cockroachdb.node --namespace "$NAMESPACE" --from-file "$NODE_CERTS_DIR"
# The ca key is not needed for any typical operations, but might be required to sign new certificates.
$UPLOAD_CA_KEY && kubectl create secret generic cockroachdb.ca.key --namespace "$NAMESPACE" --from-file "$CA_KEY_DIR"
# The ca.crt is kept in it's own secret to more easily manage cert rotation and 
# adding other operators' certificates.
kubectl create secret generic cockroachdb.ca.crt --namespace "$NAMESPACE" --from-file "$CA_KEY_DIR"
kubectl create secret generic dss.public.certs --namespace "$NAMESPACE" --from-file "$JWT_PUBLIC_CERTS_DIR"
  