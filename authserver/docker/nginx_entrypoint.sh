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
    upstream auth-server {
      server $INTERUSS_AUTH_SERVER:$INTERUSS_AUTH_PORT;
    }

    underscores_in_headers on;

    server {
      listen ${INTERUSS_AUTH_PORT_HTTPS:-8121} ssl;
      # server_name         www.example.com;
      ssl_certificate     /etc/ssl/certs/${SSL_CERT_NAME:-cert.pem};
      ssl_certificate_key /etc/ssl/private/${SSL_KEY_NAME:-key.pem};
      ssl_protocols       TLSv1 TLSv1.1 TLSv1.2;
      ssl_ciphers         HIGH:!aNULL:!MD5;

      location / {
        proxy_pass         http://auth-server;
      }
    }
  }

" > /etc/nginx/nginx.conf

echo === nginx.conf ===
cat /etc/nginx/nginx.conf
echo ==================

echo Starting nginx...

nginx -g "daemon off;"
