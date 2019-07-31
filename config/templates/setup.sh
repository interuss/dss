#!/bin/bash

# Paths to directories in which to store certificates and generated YAML files.
CLIENTS_CERTS_DIR=REPLACE_CLIENT_CERTS_DIR
NODE_CERTS_DIR=REPLACE_NODE_CERTS_DIR
DIR=REPLACE_DIRECTORY
CONTEXT=REPLACE_CONTEXT
TEMPLATES_DIR=./templates
ZONE=REPLACE_ZONE
CLUSTER_INIT=REPLACE_CLUSTER_INIT
# ------------------------------------------------------------------------------

# For each cluster create secrets containing the node and client certificates.
# Note that we create the root client certificate in both the $ZONE namespace
# and the default namespace so that its easier for clients in the default
# namespace to use without additional steps.
#
# Also create a load balancer to each clusters DNS pods.

# Delete previous secrets in case they have changed
kubectl delete secret cockroachdb.client.root --context $CONTEXT 
kubectl delete secret cockroachdb.client.root --namespace $ZONE --context $CONTEXT
kubectl delete secret cockroachdb.node --namespace $ZONE --context $CONTEXT 
# Now we can set up the certs since we can get the lbs ip address.
kubectl create secret generic cockroachdb.client.root --from-file $CLIENTS_CERTS_DIR --context $CONTEXT 
kubectl create secret generic cockroachdb.client.root --namespace $ZONE --from-file $CLIENTS_CERTS_DIR --context $CONTEXT
kubectl create secret generic cockroachdb.node --namespace $ZONE --from-file $NODE_CERTS_DIR --context $CONTEXT 

kubectl apply -f ${TEMPLATES_DIR}/dns-lb.yaml --context $CONTEXT 
kubectl apply -f ${DIR}/external-name-svc.yaml --context $CONTEXT 
# Create the cockroach resources in each cluster.
kubectl apply -f ${DIR}/cockroachdb-statefulset-secure.yaml --namespace $ZONE --context $CONTEXT 

# Clean up the node certs.
# Also clean up node certs when done
# rm ${CERTS_DIR}/node.*

# Make sure this has been run exactly once for the entire cluster.
if $CLUSTER_INIT; then
  sleep 30
  kubectl create -f ${TEMPLATES_DIR}/cluster-init-secure.yaml --namespace $ZONE --context $CONTEXT 
fi