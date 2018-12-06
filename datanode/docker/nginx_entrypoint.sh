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

echo nginx_entrypoint.sh started

echo "

  worker_processes 1;

  events { worker_connections 1024; }

  http {
    upstream storage-api {
      server $STORAGE_API_SERVER:$STORAGE_API_PORT;
    }

    underscores_in_headers on;

    server {
      listen ${INTERUSS_API_PORT_HTTP:-8120};

      listen ${INTERUSS_API_PORT_HTTPS:-8121} ssl;
      # server_name         www.example.com;
      ssl_certificate     /etc/ssl/certs/${SSL_CERT_NAME:-cert.pem};
      ssl_certificate_key /etc/ssl/private/${SSL_KEY_NAME:-key.pem};
      ssl_protocols       TLSv1 TLSv1.1 TLSv1.2;
      ssl_ciphers         HIGH:!aNULL:!MD5;

      location / {
        proxy_pass         http://storage-api;
      }
    }
  }

" > /etc/nginx/nginx.conf

echo === nginx.conf ===
cat /etc/nginx/nginx.conf
echo ==================

if sha256sum -c test_signatures --status ;
then
  echo "*****************"
  echo "*****************"
  echo "*****WARNING***** Using test SSL certificates. These should NOT be used in production environments!"
  echo "*****************"
  echo "*****************"
fi

echo Starting nginx...

nginx -g "daemon off;"
