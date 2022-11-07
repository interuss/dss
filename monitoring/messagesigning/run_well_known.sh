#!/usr/bin/env bash

docker image build --no-cache -t well-known .

docker run -p 8077:8077 --name well-known -d well-known
