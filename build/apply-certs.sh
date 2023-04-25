#!/usr/bin/env bash

set -eo pipefail

if [ "$1" == "" ]; then
  echo "usage: ./apply-certs.sh <CLUSTER_CONTEXT> <NAMESPACE>"
  exit 1
fi

if [ "$2" == "" ]; then
  echo "usage: ./apply-certs.sh <CLUSTER_CONTEXT> <NAMESPACE>"
  exit 1
fi

echo '========================================================================='
echo '= Note that errors below are acceptable as long as the terminal message ='
echo '= is success.                                                           ='
echo '========================================================================='

set -e
set -x

# Paths to directories in which to store certificates and generated YAML files.
CONTEXT="$1"
DIR="$(pwd)"
NAMESPACE="$2"

# Replace characters breaking folder names
WORKSPACE=$(echo "${CONTEXT}" | tr ':/' '_')
CLIENTS_CERTS_DIR="$DIR/workspace/$WORKSPACE/client_certs_dir"
NODE_CERTS_DIR="$DIR/workspace/$WORKSPACE/node_certs_dir"
CA_KEY_DIR="$DIR/workspace/$WORKSPACE/ca_key_dir"
CA_CRT_DIR="$DIR/workspace/$WORKSPACE/ca_certs_dir"
JWT_PUBLIC_CERTS_DIR="$DIR/jwt-public-certs"
UPLOAD_CA_KEY=true

# Delete previous secrets in case they have changed.
kubectl create namespace "$NAMESPACE"  --context "$CONTEXT" || true

kubectl delete secret cockroachdb.client.root --namespace default --context "$CONTEXT" || true
kubectl delete secret cockroachdb.client.root --namespace "$NAMESPACE"  --context "$CONTEXT" || true
kubectl delete secret cockroachdb.node --namespace "$NAMESPACE"  --context "$CONTEXT" || true
kubectl delete secret cockroachdb.ca.crt --namespace "$NAMESPACE"  --context "$CONTEXT" || true
kubectl delete secret cockroachdb.ca.key --namespace "$NAMESPACE"  --context "$CONTEXT" || true
kubectl delete secret dss.public.certs --namespace "$NAMESPACE"  --context "$CONTEXT" || true

kubectl create secret generic cockroachdb.client.root --namespace default --from-file "$CLIENTS_CERTS_DIR"  --context "$CONTEXT"
if [[ $NAMESPACE != "default" ]]; then
  kubectl create secret generic cockroachdb.client.root --namespace "$NAMESPACE" --from-file "$CLIENTS_CERTS_DIR"  --context "$CONTEXT"
fi
kubectl create secret generic cockroachdb.node --namespace "$NAMESPACE" --from-file "$NODE_CERTS_DIR"  --context "$CONTEXT"
# The ca key is not needed for any typical operations, but might be required to sign new certificates.
$UPLOAD_CA_KEY && kubectl create secret generic cockroachdb.ca.key --namespace "$NAMESPACE" --from-file "$CA_KEY_DIR"  --context "$CONTEXT"
# The ca.crt is kept in it's own secret to more easily manage cert rotation and
# adding other operators' certificates.
kubectl create secret generic cockroachdb.ca.crt --namespace "$NAMESPACE" --from-file "$CA_CRT_DIR"  --context "$CONTEXT"
kubectl create secret generic dss.public.certs --namespace "$NAMESPACE" --from-file "$JWT_PUBLIC_CERTS_DIR"  --context "$CONTEXT"

echo '========================================================================='
echo '= Secrets uploaded successfully.                                        ='
echo '========================================================================='
