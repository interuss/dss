#!/bin/sh

# This startup script is meant to be invoked from within a Docker container
# started by docker-compose_dss.yaml, not on a local system.
#
# It forms the Raft groups of the core services: the first one bootstraps them, then, with the with-raft-cluster
# profile, the others are added to them one at a time.

set -e

# The admin socket of each core service, then the URL at which the others reach it.
NODES="/run/dss/admin.sock=http://local-dss-core-service:9011"
if [ "${COMPOSE_PROFILES#*"with-raft-cluster"}" != "${COMPOSE_PROFILES}" ]; then
  # If the with-raft-cluster profile is enabled, add the two other core services to the list of nodes.
  NODES="${NODES} /run/dss2/admin.sock=http://local-dss-core-service-2:9011 /run/dss3/admin.sock=http://local-dss-core-service-3:9011"
fi

first_socket=""
for node in ${NODES}; do
  socket="${node%%=*}"
  url="${node#*=}"
  if [ -z "${first_socket}" ]; then
    first_socket="${socket}"
  fi

  # The admin API is served as soon as the core service runs, retry until then.
  nodes=""
  for _ in $(seq 1 120); do
    if nodes=$(dssctl raft node --admin_socket "${socket}" 2>/dev/null); then
      break
    fi
    nodes=""
    sleep 1
  done
  if [ -z "${nodes}" ]; then
    echo "Error: the admin API at ${socket} is not reachable." >&2
    exit 1
  fi

  # The groups in which the core service is waiting: it is already a member of the others, e.g. after a restart.
  groups=$(echo "${nodes}" | awk '$2 == "waiting" {sub(":", "", $1); print $1}' | paste -s -d, -)
  if [ -z "${groups}" ]; then
    continue
  fi
  if [ "${socket}" = "${first_socket}" ]; then
    echo "Bootstrapping the ${groups} Raft groups..."
    dssctl raft bootstrap --admin_socket "${first_socket}" --groups "${groups}"
  else
    echo "Adding ${url} to the ${groups} Raft groups..."
    dssctl raft add --admin_socket "${first_socket}" --url "${url}" --groups "${groups}"
  fi
done
echo "Raft groups formed."
