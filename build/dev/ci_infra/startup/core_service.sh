#!/bin/sh
# Temporary copy of interuss/monitoring build/dev, used by CI to test a multi-node raft cluster
# while we define the testing stack for the raftstore. Keep aligned with monitoring (see README.md).
# shellcheck disable=SC2086

set -e

# This startup script is meant to be invoked from within a Docker container
# started by docker-compose.yaml, not on a local system.

DEBUG_ON=${1:-0}
JWT_AUDIENCES="localhost,host.docker.internal,dss.lb.localutm,${JWT_AUDIENCES}"

# apply netem config for intra/inter-USS subnets, if requested
if [ -n "$INTRA_USS_NETEM_CONF" ] || [ -n "$INTER_USS_NETEM_CONF" ]; then
  apk add iproute2-tc

  # Get the first two bytes of the address (/16)
  NETEM_NET_PREFIX=$(echo "$INTER_USS_SUBNET" | cut -d. -f1-2)
  # List IP addresses to find the correct interface
  NETEM_IFACE=$(ip -o -4 addr show | grep -F " inet ${NETEM_NET_PREFIX}." | head -n 1 | awk '{print $2}')
  if [ -z "$NETEM_IFACE" ]; then
    echo "ERROR: no interface found in subnet ${INTER_USS_SUBNET}, refusing to start without traffic shaping" >&2
    exit 1
  fi
  echo "Applying netem on interface ${NETEM_IFACE}"

  # create handle on the USS network interface
  tc qdisc add dev "$NETEM_IFACE" root handle 1: prio

  if [ -n "$INTRA_USS_NETEM_CONF" ]; then
    tc qdisc add dev "$NETEM_IFACE" parent 1:2 handle 30: netem $INTRA_USS_NETEM_CONF
    tc filter add dev "$NETEM_IFACE" parent 1:0 protocol ip prio 1 u32 match ip dst "$INTRA_USS_SUBNET" flowid 1:2
  fi

  if [ -n "$INTER_USS_NETEM_CONF" ]; then
    tc qdisc add dev "$NETEM_IFACE" parent 1:3 handle 31: netem $INTER_USS_NETEM_CONF
    tc filter add dev "$NETEM_IFACE" parent 1:0 protocol ip prio 2 u32 match ip dst "$INTER_USS_SUBNET" flowid 1:3
  fi
fi

DATASTORE_CONNECTION="-store_type raft -raft_node_id=${RAFT_ID} -rid_raft_peers=${RID_RAFT_NODES} -scd_raft_peers=${SCD_RAFT_NODES} -aux_raft_peers=${AUX_RAFT_NODES} -raft_datadir /raftdata"

if [ "$DEBUG_ON" = "1" ]; then
  echo "Debug Mode: on"

  # Linter is disabled to properly unwrap $DATASTORE_CONNECTION.
  # shellcheck disable=SC2086
  dlv --headless --listen=:4000 --api-version=2 --accept-multiclient exec --continue /usr/bin/core-service -- \
  ${DATASTORE_CONNECTION} \
  -public_key_files /var/test-certs/auth2.pem \
  -log_format console \
  -dump_requests \
  -addr :80 \
  -accepted_jwt_audiences ${JWT_AUDIENCES} \
  -enable_scd \
  -allow_http_base_urls \
  -locality local_dev \
  -public_endpoint http://127.0.0.1:80 \
  ${CORE_SERVICE_EXTRA_FLAGS}
else
  echo "Debug Mode: off"

  # Linter is disabled to properly unwrap $DATASTORE_CONNECTION.
  # shellcheck disable=SC2086
  /usr/bin/core-service \
  ${DATASTORE_CONNECTION} \
  -public_key_files /var/test-certs/auth2.pem \
  -log_format console \
  -dump_requests \
  -addr :80 \
  -accepted_jwt_audiences ${JWT_AUDIENCES} \
  -enable_scd \
  -allow_http_base_urls \
  -locality local_dev \
  -public_endpoint http://127.0.0.1:80 \
  ${CORE_SERVICE_EXTRA_FLAGS}
fi
