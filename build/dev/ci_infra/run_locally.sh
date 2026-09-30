#!/usr/bin/env bash
# Temporary copy of interuss/monitoring build/dev, used by CI to test a multi-node raft cluster
# while we define the testing stack for the raftstore. Keep aligned with monitoring (see README.md).

set -eo pipefail

# This script deploys a local UTM interoperability ecosystem.
# For full documentation on parameters, usage, sizing limits, and troubleshooting, see build/dev/README.md in interuss/monitoring.

if [[ -z $(command -v docker) ]]; then
  echo "docker is required but not installed.  Visit https://docs.docker.com/install/ to install."
  exit 1
fi

OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi

cd "${BASEDIR}" || exit 1

export NUM_USS=${NUM_USS:-2}
NUM_NODES=${NUM_NODES:-1}

export DSS_IMAGE="${DSS_IMAGE:-interuss-local/dss}"
export CORE_SERVICE_EXTRA_FLAGS="${CORE_SERVICE_EXTRA_FLAGS-}"

DC_COMMAND=$*

if [[ ! "$DC_COMMAND" ]]; then
  DC_COMMAND="up"
  DC_OPTIONS="--build -d"
elif [[ "$DC_COMMAND" == "down" ]]; then
  DC_OPTIONS="--volumes --remove-orphans"
elif [[ "$DC_COMMAND" == "debug" ]]; then
  DC_COMMAND=up
  DC_OPTIONS="-d"
  export DEBUG_ON=1
fi

if [[ "$DC_COMMAND" == up* ]]; then
  DC_COMMAND=${DC_COMMAND//--wait/}
  if [[ ! "$DC_COMMAND" =~ "-d" && ! "$DC_OPTIONS" =~ "-d" ]]; then
    DC_OPTIONS="${DC_OPTIONS} -d"
  fi
  echo "Creating networks..."
  docker network create --subnet=172.27.0.0/16 \
                        --ip-range=172.27.0.0/24 \
                        --gateway=172.27.0.1 \
                        dss_internal_network || true
  docker network create interop_ecosystem_network || true
  echo "Starting containers..."
fi

if [[ "$DC_COMMAND" == up* ]]; then
  RID_RAFT_NODES=""
  SCD_RAFT_NODES=""
  AUX_RAFT_NODES=""
  for ((i=1; i<=NUM_USS; i++)); do
    for ((j=1; j<=NUM_NODES; j++)); do
      NODE_IDX=$(( (i-1) * NUM_NODES + j ))
      PADDED_NODE_IDX=$(printf "%02d" "$NODE_IDX")
      NODE_IP="172.27.${i}.${j}"
      RID_RAFT_NODES="${RID_RAFT_NODES},${NODE_IDX}=http://${NODE_IP}:95${PADDED_NODE_IDX}"
      SCD_RAFT_NODES="${SCD_RAFT_NODES},${NODE_IDX}=http://${NODE_IP}:96${PADDED_NODE_IDX}"
      AUX_RAFT_NODES="${AUX_RAFT_NODES},${NODE_IDX}=http://${NODE_IP}:97${PADDED_NODE_IDX}"
    done
  done
  export RID_RAFT_NODES=${RID_RAFT_NODES#,}
  export SCD_RAFT_NODES=${SCD_RAFT_NODES#,}
  export AUX_RAFT_NODES=${AUX_RAFT_NODES#,}

  echo "=== Local DSS pool configuration ==="
  echo "NUM_USS=${NUM_USS}"
  echo "NUM_NODES=${NUM_NODES}"
  echo "DSS_IMAGE=${DSS_IMAGE}"
  echo "CORE_SERVICE_EXTRA_FLAGS=${CORE_SERVICE_EXTRA_FLAGS}"
  echo "INTER_USS_NETEM_CONF=${INTER_USS_NETEM_CONF}"
  echo "INTRA_USS_NETEM_CONF=${INTRA_USS_NETEM_CONF}"
  echo "DC_COMMAND=${DC_COMMAND}"
  echo "===================================="

  for ((i=1; i<=NUM_USS; i++)); do
    for ((j=1; j<=NUM_NODES; j++)); do
      export USS_IDX=$i
      export USS_NODE_IDX=$j
      NODE_IDX=$(( (i-1) * NUM_NODES + j ))
      export RAFT_ID=$NODE_IDX
      PADDED_NODE_IDX=$(printf "%02d" "$NODE_IDX")
      export PADDED_NODE_IDX

      export COMPOSE_PROFILES=
      if [ "$i" -eq 1 ] && [ "$j" -eq 1 ]; then
        export COMPOSE_PROFILES=oauth,lb
      fi

      # e.g. for USS 3 node 2 the IP would be 172.27.3.2
      export DSS_IP="172.27.$USS_IDX.$USS_NODE_IDX"

      # shellcheck disable=SC2086
      docker compose -f docker-compose.yaml -p "ci_infra_${USS_IDX}-${USS_NODE_IDX}" $DC_COMMAND $DC_OPTIONS &
      sleep 0.1 # reduce probability of race condition in joining network at container start
    done
  done
  wait
else
  PROJECTS=$(docker compose ls -a -q 2>/dev/null | grep '^ci_infra_' || true)
  if [[ -n "$PROJECTS" ]]; then
    echo "=== Local DSS pool: executing '$DC_COMMAND' on discovered projects ==="
    echo "$PROJECTS"
    echo "======================================================================="
    for p in $PROJECTS; do
      if [[ "$p" =~ ^ci_infra_([0-9]+)-([0-9]+)$ ]]; then
        export USS_IDX="${BASH_REMATCH[1]}"
        export USS_NODE_IDX="${BASH_REMATCH[2]}"
        export DSS_IP="172.27.$USS_IDX.$USS_NODE_IDX"
        export PADDED_NODE_IDX
        PADDED_NODE_IDX="$(printf "%02d" "$USS_NODE_IDX")"
        export NUM_USS=1
        export DSS_IMAGE="${DSS_IMAGE:-interuss-local/dss}"
        export COMPOSE_PROFILES="*"

        # shellcheck disable=SC2086
        docker compose -f docker-compose.yaml --profile "*" -p "$p" $DC_COMMAND $DC_OPTIONS &
      fi
    done
    wait
  else
    echo "No deployed local DSS projects found to execute '$DC_COMMAND'."
  fi
fi

if [[ "$DC_COMMAND" == up* ]]; then
  echo "Verifying and repairing docker network connections..."

  check_and_connect() {
    local container=$1
    local network=$2
    if docker ps -a --format '{{.Names}}' | grep -q "^${container}$"; then
      if ! docker inspect "${container}" --format '{{json .NetworkSettings.Networks}}' | grep -q "\"${network}\""; then
        echo "Warning: Container ${container} is not connected to ${network}. Reconnecting and restarting so the entrypoint reapplies traffic shaping..."
        docker network connect "${network}" "${container}" || {
          docker stop -t 2 "${container}" >/dev/null 2>&1 || true
          docker network connect "${network}" "${container}"
        }
        docker restart "${container}"
      fi
    fi
  }

  for ((i=1; i<=NUM_USS; i++)); do
    for ((j=1; j<=NUM_NODES; j++)); do
      check_and_connect "ci_infra_${i}-${j}-dss-1" "dss_internal_network"
      check_and_connect "ci_infra_${i}-${j}-dss-1" "interop_ecosystem_network"
    done
  done

  check_and_connect "ci_infra_1-1-oauth-1" "interop_ecosystem_network"
  echo "Network verification complete."

  echo "Waiting for all containers to become healthy..."
  timeout=240
  interval=5
  elapsed=0
  all_healthy=true

  while [ $elapsed -lt $timeout ]; do
    all_healthy=true
    unhealthy_containers=()

    check_container_health() {
      local container=$1
      if ! docker ps --format '{{.Names}}' | grep -q "^${container}$"; then
        all_healthy=false
        unhealthy_containers+=("$container (not running)")
        return
      fi

      local health_status
      health_status=$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$container")
      if [[ "$health_status" == "unhealthy" || "$health_status" == "starting" ]]; then
        all_healthy=false
        unhealthy_containers+=("$container ($health_status)")
      fi
    }

    for ((i=1; i<=NUM_USS; i++)); do
      for ((j=1; j<=NUM_NODES; j++)); do
        check_container_health "ci_infra_${i}-${j}-dss-1"
      done
    done
    check_container_health "ci_infra_1-1-oauth-1"

    if [ "$all_healthy" = true ]; then
      echo "All containers are healthy!"
      break
    fi

    echo "Still waiting (elapsed ${elapsed}s)... Unhealthy/starting/not-running:"
    for uc in "${unhealthy_containers[@]}"; do
      echo "  - $uc"
    done

    sleep $interval
    elapsed=$((elapsed + interval))
  done

  if [ "$all_healthy" != true ]; then
    echo "Error: Timeout waiting for containers to become healthy."
    exit 1
  fi
fi

if [[ "$DC_COMMAND" == "down" ]]; then
  echo "Removing networks..."
  docker network rm dss_internal_network 2>/dev/null || true
  docker network rm interop_ecosystem_network 2>/dev/null || true
fi
