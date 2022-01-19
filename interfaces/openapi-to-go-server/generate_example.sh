#!/usr/bin/env bash

OS=$(uname)
if [[ $OS == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi

cd ${BASEDIR}

if [[ ! -f "$FILE" ]]; then
    curl https://raw.githubusercontent.com/astm-utm/Protocol/v1.0.0/utm.yaml > utm.yaml
fi

docker image build -t openapi-to-go-server .

docker container run -it \
  	-v $(pwd):/resources/in \
	  -v $(pwd)/example:/resources/out \
	  openapi-to-go-server \
	      --input_yaml /resources/in/utm.yaml \
	      --path_prefix /scd \
	      --output_folder /resources/out \
	      --include_endpoint_tags dss \
	      --package main \
	      --include_types \
	      --include_interface \
	      --include_server \
	      --include_common \
	      --include_example
