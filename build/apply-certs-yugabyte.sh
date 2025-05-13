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
CLIENTS_CERTS_DIR="$DIR/workspace-yugabyte/$WORKSPACE/client_certs_dir"
MASTER_CERTS_DIR="$DIR/workspace-yugabyte/$WORKSPACE/master_certs_dir"
TSERVER_CERTS_DIR="$DIR/workspace-yugabyte/$WORKSPACE/tserver_certs_dir"
# CA_KEY_DIR="$DIR/workspace/$WORKSPACE/ca_key_dir"
# CA_CRT_DIR="$DIR/workspace/$WORKSPACE/ca_certs_dir"
JWT_PUBLIC_CERTS_DIR="$DIR/jwt-public-certs"

# Delete previous secrets in case they have changed.
kubectl create namespace "$NAMESPACE"  --context "$CONTEXT" || true

kubectl delete secret yb-master-yugabyte-tls-cert --namespace "$NAMESPACE"  --context "$CONTEXT" || true
kubectl delete secret yb-tserver-yugabyte-tls-cert --namespace "$NAMESPACE"  --context "$CONTEXT" || true
kubectl delete secret yugabyte-tls-client-cert --namespace "$NAMESPACE"  --context "$CONTEXT" || true
kubectl delete secret dss.public.certs --namespace "$NAMESPACE"  --context "$CONTEXT" || true

kubectl create secret generic yb-master-yugabyte-tls-cert --namespace "$NAMESPACE" --from-file "$MASTER_CERTS_DIR"  --context "$CONTEXT"
kubectl create secret generic yb-tserver-yugabyte-tls-cert --namespace "$NAMESPACE" --from-file "$TSERVER_CERTS_DIR"  --context "$CONTEXT"
kubectl create secret generic yugabyte-tls-client-cert --namespace "$NAMESPACE" --from-file "$CLIENTS_CERTS_DIR"  --context "$CONTEXT"


kubectl create secret generic dss.public.certs --namespace "$NAMESPACE" --from-file "$JWT_PUBLIC_CERTS_DIR"  --context "$CONTEXT"

echo '========================================================================='
echo '= Secrets uploaded successfully.                                        ='
echo '========================================================================='
