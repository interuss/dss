#!/bin/sh
# Temporary copy of interuss/monitoring build/dev, used by CI to test a multi-node raft cluster
# while we define the testing stack for the raftstore. Keep aligned with monitoring (see README.md).
set -e

cat > /tmp/haproxy.cfg <<'EOF'
global
    maxconn 1024

defaults
    mode http
    timeout connect 5s
    timeout client 30s
    timeout server 30s

resolvers docker
    nameserver dns1 127.0.0.11:53

frontend dss_in
    bind *:80
    # Same port as the standalone DSS instance, used by build/dev/*_locally.sh
    bind *:8082
    default_backend dss_pool

backend dss_pool
    balance roundrobin
EOF

i=1
while [ "$i" -le "$NUM_USS" ]; do
  j=1
  while [ "$j" -le "$NUM_NODES" ]; do
    echo "    server dss${j}_${i} dss${j}.uss${i}.localutm:80 check resolvers docker init-addr none"
    j=$((j+1))
  done
  i=$((i+1))
done >> /tmp/haproxy.cfg

exec haproxy -f /tmp/haproxy.cfg
