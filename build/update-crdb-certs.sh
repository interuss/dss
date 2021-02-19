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
CA_CRT_DIR="$DIR/workspace/$CONTEXT/ca_certs_dir"

kubectl delete secret cockroachdb.ca.crt --namespace "$NAMESPACE"  --context "$CONTEXT" || true
kubectl create secret generic cockroachdb.ca.crt --namespace "$NAMESPACE" --from-file "$CA_CRT_DIR"  --context "$CONTEXT"

echo '========================================================================='
echo '= Secrets uploaded successfully.                                        ='
echo '========================================================================='
