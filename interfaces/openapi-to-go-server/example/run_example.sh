#!/usr/bin/env bash

OS=$(uname)
if [[ $OS == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi

cd ${BASEDIR}

docker image build -t openapi-to-go-server-demo .

echo "Running server..."
docker container run -it -p 8080:8080 openapi-to-go-server-demo
