#!/bin/sh

echo "spinning up cockroachdb"
docker run -d --name dss-crdb-for-e2e-testing -p 26257:26257 -v "$(pwd)/config/test-certs/:/certs/" \
  cockroachdb/cockroach:v19.1.2 start \
    --certs-dir "/certs/cockroach-certs"  \
    --http-addr 0.0.0.0 \
    --advertise-addr=localhost:26257 \
  > /dev/null

echo "spinning up grpc-backend"
docker run -d --name dss-grpc-for-e2e-testing $(docker build -q cmds/grpc-backend/) -p 8081:8081 -v "$(pwd)/config/test-certs/:/certs/" \
    --addr=:8081 \
    --cockroach_host=localhost \
    --cockroach_port=26257 \
    --cockroach_ssl_mode=verify-full \
    --cockroach_user=root \
    --cockroach_ssl_dir="/certs/cockroach-certs" \
    --public_key_file="/certs/oauth.crt" \
    --jwt_audience=test \
  > /dev/null

echo "spinning up http-gateway"
docker run -d --name dss-http-for-e2e-testing $(docker build -q cmds/http-gateway/) -p 8082:8082 \
  --addr=:8082 \
  --grpc-backend=localhost:8081 \
 > /dev/null

echo "spinning up dummy-oauth"
docker run -d --name dss-oauth-for-e2e-testing $(docker build -q cmds/dummy-oauth/) -p 8085:8085 -v "$(pwd)/config/test-certs/:/certs/" \
  --addr=:8085 \
  --private_key_file="/certs/oauth.key" \
  &

echo "starting tests"
pytest \
    --oauth-token-endpoint  "localhost:8085" \
    --dss-endpoint "localhost:8081" \
    -vv monitoring/prober
echo "test over"

docker stop dss-crdb-for-e2e-testing > /dev/null
docker rm dss-crdb-for-e2e-testing > /dev/null
