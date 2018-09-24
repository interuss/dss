#!/bin/bash

# Copyright 2018 Google Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Allow the container to be started with `--user`
if [[ "$1" = 'zkServer.sh' && "$(id -u)" = '0' ]]; then
    chown -R "$ZOO_USER" "$ZOO_DATA_DIR" "$ZOO_DATA_LOG_DIR"
    exec su-exec "$ZOO_USER" "$0" "$@"
fi

# Generate Zookeeper config
ZOO_CONFIG="$ZOO_CONF_DIR/zoo.cfg"

echo "quorumListenOnAllIPs=TRUE" > "$ZOO_CONFIG"

echo "clientPort=$ZOO_PORT" >> "$ZOO_CONFIG"
echo "dataDir=$ZOO_DATA_DIR" >> "$ZOO_CONFIG"
echo "dataLogDir=$ZOO_DATA_LOG_DIR" >> "$ZOO_CONFIG"

echo "tickTime=$ZOO_TICK_TIME" >> "$ZOO_CONFIG"
echo "initLimit=$ZOO_INIT_LIMIT" >> "$ZOO_CONFIG"
echo "syncLimit=$ZOO_SYNC_LIMIT" >> "$ZOO_CONFIG"

echo "maxClientCnxns=$ZOO_MAX_CLIENT_CNXNS" >> "$ZOO_CONFIG"

for server in $ZOO_SERVERS; do
    if [[ $server = "server.${ZOO_MY_ID}"* ]]; then
        echo "$server" | sed 's/=[0-9.]*:/=0.0.0.0:/' >> "$ZOO_CONFIG"
    else
        echo "$server" >> "$ZOO_CONFIG"
    fi
done

# Write myid
echo "${ZOO_MY_ID:-1}" > "$ZOO_DATA_DIR/myid"

exec "$@"
