#!/bin/bash

# Paths to directories in which to store certificates and generated YAML files.
NAMESPACE=
CLUSTER_INIT=
CLIENTS_CERTS_DIR=$(pwd)/generated/$NAMESPACE/client_certs_dir
NODE_CERTS_DIR=$(pwd)/generated/$NAMESPACE/node_certs_dir
DIR=$(pwd)
CONTEXT=$(kubectl config get-contexts -o name)
TEMPLATES_DIR=$DIR/templates
# ------------------------------------------------------------------------------

# Delete previous secrets in case they have changed
kubectl delete secret cockroachdb.client.root --context $CONTEXT 
kubectl delete secret cockroachdb.client.root --namespace $NAMESPACE --context $CONTEXT
kubectl delete secret cockroachdb.node --namespace $NAMESPACE --context $CONTEXT 
# Now we can set up the certs since we can get the lbs ip address.
kubectl create secret generic cockroachdb.client.root --from-file $CLIENTS_CERTS_DIR --context $CONTEXT 
kubectl create secret generic cockroachdb.client.root --namespace $NAMESPACE --from-file $CLIENTS_CERTS_DIR --context $CONTEXT
kubectl create secret generic cockroachdb.node --namespace $NAMESPACE --from-file $NODE_CERTS_DIR --context $CONTEXT 