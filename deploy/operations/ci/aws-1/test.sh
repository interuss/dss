#!/usr/bin/env bash

set -eo pipefail

# Find and change to repo root directory
OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi
cd "${BASEDIR}" || exit 1

# Initialize terraform
terraform init
# TODO: Fail if env is not clean

# Deploy the Kubernetes cluster
terraform apply -auto-approve
KUBE_CONTEXT="$(terraform output -raw cluster_context)"
WORKSPACE_LOCATION="$(terraform output -raw workspace_location)"

# Login into the Kubernetes Cluster
cd "${WORKSPACE_LOCATION}"
./get-credentials.sh
aws sts get-caller-identity

# Allow access to the cluster to AWS admins
kubectl apply -f "aws_auth_config_map.yml"

# Generate cockroachdb certificates
./make-certs.sh
./apply-certs.sh

# Install the DSS using the helm chart
cd "$BASEDIR/../../../services/helm-charts/dss"
RELEASE_NAME="dss"
helm dep update --kube-context="$KUBE_CONTEXT"
helm upgrade --install --debug --kube-context="$KUBE_CONTEXT" -f "${WORKSPACE_LOCATION}/helm_values.yml" "$RELEASE_NAME" .
kubectl wait --for=condition=complete --timeout=3m job/rid-schema-manager-1

# Test the deployment of the DSS
kubectl apply -f test/test-resources.yaml
kubectl create secret generic -n tests dummy-oauth-certs --from-file="$BASEDIR/../../../../build/test-certs/auth2.key"
kubectl wait -n tests --for=condition=complete --timeout=10m job.batch/uss-qualifier
# dummy-oauth-certs secret is deleted with the namespace using the command below
kubectl delete -f test/test-resources.yaml


if [ -n "$DO_NOT_DESTROY" ]; then
  echo "Destroy disabled. Exit."
  exit 0
fi

# Cleanup
# Delete workloads
helm uninstall --debug --kube-context="$KUBE_CONTEXT" --wait --timeout 5m "$RELEASE_NAME"

# Delete PVC to delete persistent volumes
kubectl delete pvc --wait --all=true
kubectl delete pv --wait --all=true
# TODO: Check completeness

# Delete cluster
cd "$BASEDIR"
terraform destroy -auto-approve


