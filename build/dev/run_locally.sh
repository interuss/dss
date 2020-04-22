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
	BASEDIR="$(dirname $0)/../.."
else
	BASEDIR=$(readlink -e "$(dirname "$0")/../..")
fi

cd "${BASEDIR}" || exit 1

pwd

# Make sure the necessary ports are not in use already

pinfo_8080=$(lsof -i tcp:8080)
if [ ! -z "${pinfo_8080}" ]; then
  echo "Cannot start CockroachDB because a process has port 8080 open:"
  echo ${pinfo_8080}
  exit 1
fi

pinfo_8081=$(lsof -i tcp:8081)
if [ ! -z "${pinfo_8081}" ]; then
  echo "Cannot start grpc-backend because a process has port 8081 open:"
  echo ${pinfo_8081}
  exit 1
fi

pinfo_8082=$(lsof -i tcp:8082)
if [ ! -z "${pinfo_8082}" ]; then
  echo "Cannot start http-gateway because a process has port 8082 open:"
  echo ${pinfo_8082}
  exit 1
fi

pinfo_8085=$(lsof -i tcp:8085)
if [ ! -z "${pinfo_8085}" ]; then
  echo "Cannot start dummy OAuth because a process has port 8085 open:"
  echo ${pinfo_8085}
  exit 1
fi

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
$(ps -p ${pid_grpc} > /dev/null)
grpc_ok=$?

echo "=== Starting http-gateway on :8082 ==="
go run cmds/http-gateway/main.go -grpc-backend localhost:8081 -addr :8082 &
pid_http=$!
sleep 5
$(ps -p ${pid_http} > /dev/null)
http_ok=$?

echo "=== Starting dummy OAuth server on :8085 ==="
go run cmds/dummy-oauth/main.go -private_key_file build/test-certs/auth2.key &
pid_oauth=$!
sleep 5
$(ps -p ${pid_oauth} > /dev/null)
oauth_ok=$?

if [ ${grpc_ok} -eq 0 ] && [ ${http_ok} -eq 0 ] && [ ${oauth_ok} -eq 0 ]
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
    if [ ${grpc_ok} -ne 0 ]; then
        echo "Error starting grpc-backend."
    fi
    if [ ${http_ok} -ne 0 ]; then
        echo "Error starting http-gateway."
    fi
    if [ ${oauth_ok} -ne 0 ]; then
        echo "Error starting dummy OAuth server."
    fi
    echo "Waiting on processes ${pid_grpc} (grpc), ${pid_http} (http), and ${pid_oauth} (oauth) to terminate..."
fi

wait $pid_grpc && wait $pid_http && wait $pid_oauth
echo "Stopping CockroachDB container..."
docker stop dss-crdb-for-debugging
