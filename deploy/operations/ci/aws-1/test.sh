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

terraform init
# TODO: Fail if env is not clean

## Deploy the Kubernetes cluster
terraform apply -auto-approve
KUBE_CONTEXT="$(terraform output -raw cluster_context)"
WORKSPACE_LOCATION="$(terraform output -raw workspace_location)"

cd "${WORKSPACE_LOCATION}"
./get-credentials.sh
aws sts get-caller-identity

# Allow access to the cluster to AWS admins
kubectl apply -f "aws_auth_config_map.yml"

## Generate cockroachdb certificates
./make-certs.sh
./apply-certs.sh

cd "$BASEDIR/../../../services/helm-charts/dss"
RELEASE_NAME="dss"
helm dep update --kube-context="$KUBE_CONTEXT"
helm upgrade --install --kube-context="$KUBE_CONTEXT" -f "${WORKSPACE_LOCATION}/helm_values.yml" "$RELEASE_NAME" .

# TODO: Test the deployment of the DSS

if [ -n "$DO_NOT_DESTROY" ]; then
  "Destroy disabled. Exit."
  exit 0
fi

# Cleanup
# Delete workloads
helm uninstall --kube-context="$KUBE_CONTEXT" "$RELEASE_NAME"

# Delete PVC to delete persistant volumes
kubectl delete pvc --all=true

# Delete cluster
cd "$BASEDIR"
terraform destroy -auto-approve


