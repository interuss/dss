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

# This file may be adapted to bring up or down a production node in one
# command. Replace content in ** ** characters with your server-specific
# information, and see README.md for more information.

if [ "$1" == "" ]; then
  echo "You must pass in the server ID (1-999)"
  exit 0
fi
echo "Server ID: $1"

export SSL_CERT_PATH=**FULL LOCAL PATH CONTAINING SSL CERTIFICATE**
export SSL_KEY_PATH=**FULL LOCAL PATH CONTAINING SSL CERTIFICATE KEY**
export SSL_CERT_NAME=**NAME OF SSL CERTIFICATE FILE (e.g., cert.crt or cert.chained.pem)**
export SSL_KEY_NAME=**NAME OF SSL KEY FILE (e.g., private.pem or cert.key)**

echo "SSL cert ${SSL_CERT_PATH}/${SSL_KEY_NAME}"
echo "SSL key ${SSL_KEY_PATH}/${SSL_KEY_NAME}"

export ZOO_MY_ID=$1
export INTERUSS_PUBLIC_KEY=**PUBLIC KEY OF YOUR AUTH SERVER**
export ZOO_SERVERS=**YOUR INTERUSS PLATFORM NETWORK CONFIGURATION**
export ZOO_SERVERS=`echo $ZOO_SERVERS | sed "s/server\.${ZOO_MY_ID}=[^:]*:/server.${ZOO_MY_ID}=0.0.0.0:/"`

echo "ZOO_SERVERS=${ZOO_SERVERS}"

if ! command -v docker-compose; then
  echo "No docker-compose found; using docker-compose container alias"
  shopt -s expand_aliases
  alias docker-compose='printenv > .env; docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v "$PWD:/rootfs/$PWD" -w="/rootfs/$PWD" docker/compose:1.23.1'
fi

echo "Downloading the latest deployment files..."
wget -N https://raw.githubusercontent.com/wing-aviation/InterUSS-Platform/tcl4/datanode/docker/docker-compose.yaml
wget -N https://raw.githubusercontent.com/wing-aviation/InterUSS-Platform/tcl4/datanode/docker/docker-compose_localssl.yaml

echo "Stopping any existing docker containers..."
docker-compose -f docker-compose.yaml -f docker-compose_localssl.yaml -p datanode down

echo "Downloading the latest images..."
docker pull interussplatform/reverse_proxy
docker pull interussplatform/storage_api:tcl4

echo "Starting the new docker containers..."
docker-compose -f docker-compose.yaml -f docker-compose_localssl.yaml -p datanode up -d
