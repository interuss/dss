#!/usr/bin/env bash

OS=$(uname)
if [[ $OS == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi

cd "${BASEDIR}" || exit

docker image build -t interuss/openapi-to-go-server .

docker container run -it \
  	-v "$(pwd)/../brutm/utm/utm.yaml:/resources/utm.yaml" \
  	-v "$(pwd)/../brutm/remoteid/remoteid.yaml:/resources/rid.yaml" \
	  -v "$(pwd)/example:/resources/example" \
	  interuss/openapi-to-go-server \
	      --api_import example/api \
	      --api /resources/utm.yaml#dss@scd \
	      --api /resources/rid.yaml#dss@rid \
	      --api_folder /resources/example/api \
	      --example_folder /resources/example
