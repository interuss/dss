#!/bin/bash

# This script will deploy a standalone DSS instance via local processes.  See
# standalone_instance.md for more information.

if [[ ! -z $(docker container ls --filter name=dss-crdb-for-debugging --quiet) ]]; then
  echo "=== Cleaning up pre-existing CockroachDB container... ==="
  docker stop dss-crdb-for-debugging
fi

OS=$(uname)
if [[ $OS == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname $0)/.."
else
	BASEDIR=$(readlink -e "$(dirname "$0")/../..")
fi

cd "${BASEDIR}" || exit 1

pwd

echo "=== Starting CockroachDB with admin port on :8080... ==="
docker run -d --rm --name dss-crdb-for-debugging -p 26257:26257 -p 8080:8080  cockroachdb/cockroach:v19.1.2 start --insecure > /dev/null

sleep 5
echo "=== Starting grpc-backend on :8081 ==="
go run cmds/grpc-backend/main.go \
    -cockroach_host localhost \
    -public_key_file build/test-certs/auth2.pem \
    -reflect_api \
    -log_format console \
    -dump_requests \
    -accepted_jwt_audiences localhost &
pid_grpc=$!

sleep 5
echo "=== Starting http-gateway on :8082 ==="
go run cmds/http-gateway/main.go -grpc-backend localhost:8081 -addr :8082 &
pid_http=$!

sleep 5
echo "=== Starting dummy OAuth server on :8085 ==="
go run cmds/dummy-oauth/main.go -private_key_file build/test-certs/auth2.key &
pid_oauth=$!


if [ -d "/proc/${pid_grpc}" ] && [ -d "/proc/${pid_http}" ] && [ -d "/proc/${pid_oauth}" ]
then
    echo "=============================================================="
    echo "All systems GO; DSS instance is running locally!"
    echo "  CockroachDB container `docker container ls --filter name=dss-crdb-for-debugging --quiet`"
    echo "  grpc-backend PID ${pid_grpc}"
    echo "  http-gateway PID ${pid_http}"
    echo "  Dummy OAuth PID ${pid_oauth}"
    echo "Dummy OAuth Server Address:      http://localhost:8085/token"
    echo "DSS HTTP Gateway Server Address: http://localhost:8082/healthy"
    echo
    echo "Run ./check_dss.sh from another terminal to verify minimal DSS functionality"
    echo "Press ctrl-c here to stop the DSS"
    echo "=============================================================="
else
    echo "Processes did not start correctly."
fi


wait $pid_grpc && wait $pid_http && wait $pid_oauth
echo "Stopping CockroachDB container..."
docker stop dss-crdb-for-debugging
