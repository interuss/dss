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

# This file may be adapted to bring up or down a production auth server in one
# command. Replace content in ** ** characters with your server-specific
# information, and see README.md for more information.

if [ "$1" == "" ]; then
  echo "You must specify a command ('up' or 'down')"
  exit 0
fi

wget -N https://raw.githubusercontent.com/wing-aviation/InterUSS-Platform/publicportal/authserver/docker/docker-compose.yaml
export INTERUSS_AUTH_PATH=**FULL LOCAL PATH CONTAINING roster.txt, public.pem, and private.pem**
export SSL_CERT_PATH=**FULL LOCAL PATH CONTAINING SSL CERTIFICATE**
export SSL_KEY_PATH=**FULL LOCAL PATH CONTAINING SSL CERTIFICATE KEY**
export SSL_CERT_NAME=**NAME OF SSL CERTIFICATE FILE (e.g., cert.crt or cert.chained.pem)**
export SSL_KEY_NAME=**NAME OF SSL KEY FILE (e.g., private.pem or cert.key)**
export INTERUSS_AUTH_ISSUER=**YOUR ISSUING DOMAIN NAME (e.g., auth.staging.interussplatform.com)**
if [ "$1" == "up" ]; then
  docker-compose -p authserver up -d
else
  docker-compose -p authserver down
fi
