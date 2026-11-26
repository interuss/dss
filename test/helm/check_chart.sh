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

helm dependency update "${CHART}"
helm lint "${CHART}" -f "${EXAMPLE}"

for provider in google aws minikube; do
	for datastore in cockroachdb yugabyte raft; do
		helm template dss "${CHART}" -f "${EXAMPLE}" \
			--set "global.cloudProvider=${provider}" \
			--set "cockroachdb.enabled=$([[ ${datastore} == cockroachdb ]] && echo true || echo false)" \
			--set "yugabyte.enabled=$([[ ${datastore} == yugabyte ]] && echo true || echo false)" \
			--set "raft.enabled=$([[ ${datastore} == raft ]] && echo true || echo false)" \
			>/dev/null
		echo "Rendered: cloudProvider=${provider} datastore=${datastore}"
	done
done
