#!/bin/bash
echo "cleaning up any pre-existing containers"
docker stop dss-crdb-for-debugging

echo "starting cockroachdb with admin port on :8080"
docker run -d --rm --name dss-crdb-for-debugging -p 26257:26257 -p 8080:8080  cockroachdb/cockroach:v19.1.2 start --insecure > /dev/null

sleep 5
echo "starting grpc backend on :8081"
go run cmds/grpc-backend/main.go -cockroach_host localhost -public_key_file config/oauth.pem -reflect_api true &
pid1=$!

sleep 5
echo "starting http-gateway on :8082"
go run cmds/http-gateway/main.go -grpc-backend=localhost:8081 -addr :8082 &
pid2=$!

wait $pid1 && wait $pid2
docker stop dss-crdb-for-debugging
