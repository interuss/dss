#!/usr/bin/env bash

# Lints the DSS Helm chart and renders it for every datastore and cloud provider,
# so that a change to the chart, its schema or its example values that breaks a
# documented configuration fails before release.
#
# Requires helm (see deploy/services/helm-charts/dss/README.md) and network
# access to download the chart dependencies.

set -eo pipefail

OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi
cd "${BASEDIR}/../.." || exit 1

CHART="deploy/services/helm-charts/dss"
EXAMPLE="${CHART}/values.example.yaml"
HELM_IMAGE="${HELM_IMAGE:-alpine/helm:3.22.1}"

helm_cmd() {
	docker run --rm \
		-u "$(id -u):$(id -g)" \
		-e HELM_CACHE_HOME=/tmp/.helm/cache \
		-e HELM_CONFIG_HOME=/tmp/.helm/config \
		-e HELM_DATA_HOME=/tmp/.helm/data \
		-v "$(pwd):/workspace" \
		-w /workspace \
		"${HELM_IMAGE}" "$@"
}

helm_cmd dependency update "${CHART}"

for provider in google aws minikube; do
	for datastore in cockroachdb yugabyte raft; do
		# TODO: Use an example file per datastore
		args=(-f "${EXAMPLE}"
			--set "global.cloudProvider=${provider}"
			--set "cockroachdb.enabled=$([[ ${datastore} == cockroachdb ]] && echo true || echo false)"
			--set "yugabyte.enabled=$([[ ${datastore} == yugabyte ]] && echo true || echo false)"
			--set "raft.enabled=$([[ ${datastore} == raft ]] && echo true || echo false)")
		helm_cmd lint "${CHART}" "${args[@]}" >/dev/null
		helm_cmd template dss "${CHART}" "${args[@]}" >/dev/null
		echo "Rendered: cloudProvider=${provider} datastore=${datastore}"
	done
done
