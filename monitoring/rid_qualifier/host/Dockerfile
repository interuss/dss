FROM python:3.8
# Not -alpine because: https://stackoverflow.com/a/58028091/651139
# `docker build` should be run from `monitoring` (the parent folder of this folder)
RUN apt-get update && apt-get install openssl && apt-get install ca-certificates
RUN mkdir -p /tmp/rid-host-input-files
RUN mkdir -p /app/monitoring/monitorlib
RUN mkdir -p /app/monitoring/rid_qualifier/host
COPY monitorlib/requirements.txt /app/monitoring/monitorlib/requirements.txt
COPY rid_qualifier/requirements.txt /app/monitoring/rid_qualifier/requirements.txt
COPY rid_qualifier/host/requirements.txt /app/monitoring/rid_qualifier/host/requirements.txt
WORKDIR /app/monitoring/rid_qualifier/host
RUN pip install -r requirements.txt
RUN rm -rf __pycache__
ADD . /app/monitoring
ENV PYTHONPATH /app
ARG version
ENV CODE_VERSION=$version

ENTRYPOINT []
