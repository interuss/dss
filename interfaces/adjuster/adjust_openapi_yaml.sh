#!/usr/bin/env bash

DIR_IN="$(cd "$(dirname "$1")" || exit; pwd -P)/$(basename "$1")"
DIR_OUT="$(cd "$(dirname "$2")" || exit; pwd -P)/$(basename "$2")"

OS=$(uname)
if [[ $OS == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
	DIR_IN="$(dirname "$DIR_IN")"
	DIR_OUT="$(dirname "$DIR_OUT")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
	DIR_IN=$(readlink -e "$(dirname "$DIR_IN")")
	DIR_OUT=$(readlink -e "$(dirname "$DIR_OUT")")
fi

docker image build "${BASEDIR}" -t interuss/adjust_openapi_yaml

docker container run -it \
	-v "$DIR_IN":/resources/in \
	-v "$DIR_OUT":/resources/out \
	interuss/adjust_openapi_yaml \
	--input_yaml  "/resources/in/${1##*/}" \
    --output_yaml "/resources/out/${2##*/}"
