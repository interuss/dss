#!/bin/bash

ROOT_DIR="$(git rev-parse --show-toplevel)"

# Ensure the logs directory exists
mkdir -p logs

PEERS="1=https://127.0.0.1:9021,2=https://127.0.0.1:9022,3=https://127.0.0.1:9023"

echo "=== Starting 3-node raft cluster ==="

for NODE_ID in 1 2 3; do
    ADDR=":808$NODE_ID"
    OUT="logs/node$NODE_ID.log"

    echo "Starting node $NODE_ID on port $ADDR..."
    TLS_ARGS="--raft_tls=ca=$ROOT_DIR/build/test-certs/raft-certs/ca.crt,cert=$ROOT_DIR/build/test-certs/raft-certs/node${NODE_ID}.crt,key=$ROOT_DIR/build/test-certs/raft-certs/node${NODE_ID}.key"

    go run "$ROOT_DIR"/cmds/core-service \
        --store_type=raft \
        --raft_node_id=$NODE_ID \
        --addr=$ADDR \
        --raft_peers=$PEERS \
        "$TLS_ARGS" \
        --accepted_jwt_audiences=dss \
        --public_key_files="$ROOT_DIR"/build/test-certs/auth2.pem > "$OUT" 2>&1 &
done