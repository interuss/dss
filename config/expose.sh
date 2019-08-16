#!/bin/bash

# Paths to directories in which to store certificates and generated YAML files.
for i in {0..2}
  do
    kubectl expose pod cockroachdb-$i --namespace $NAMESPACE --type=LoadBalancer --name=crdb-node-$i
  done
