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
  echo "You must specify a command ('up' or 'down')"
  exit 0
fi

wget -N https://raw.githubusercontent.com/wing-aviation/InterUSS-Platform/publicportal/datanode/docker/docker-compose.yaml
wget -N https://raw.githubusercontent.com/wing-aviation/InterUSS-Platform/publicportal/datanode/docker/docker-compose_localssl.yaml
wget -N https://**YOUR AUTH SERVER (e.g., auth.staging.interussplatform.com:8121)**/key
export SSL_CERT_PATH=**FULL LOCAL PATH CONTAINING SSL CERTIFICATE**
export SSL_KEY_PATH=**FULL LOCAL PATH CONTAINING SSL CERTIFICATE KEY**
export SSL_CERT_NAME=**NAME OF SSL CERTIFICATE FILE (e.g., cert.crt or cert.chained.pem)**
export SSL_KEY_NAME=**NAME OF SSL KEY FILE (e.g., private.pem or cert.key)**
export ZOO_MY_ID=1
export ZOO_SERVERS=0.0.0.0:2888:3888
export INTERUSS_PUBLIC_KEY="`cat key`"
if [ "$1" == "up" ]; then
  docker-compose -f docker-compose.yaml -f docker-compose_localssl.yaml -p datanode up -d
else
  docker-compose -f docker-compose.yaml -f docker-compose_localssl.yaml -p datanode down
fi
