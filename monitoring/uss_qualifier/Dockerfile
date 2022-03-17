FROM python:3.8
# Not -alpine because: https://stackoverflow.com/a/58028091/651139
# `docker build` should be run from `monitoring` (the parent folder of this folder)
RUN apt-get update && apt-get install openssl && apt-get install ca-certificates
RUN mkdir -p /app/monitoring/uss_qualifier
RUN mkdir /app/monitoring/monitorlib
COPY monitorlib/requirements.txt /app/monitoring/monitorlib/requirements.txt
COPY uss_qualifier/requirements.txt /app/monitoring/uss_qualifier/requirements.txt
WORKDIR /app/monitoring/uss_qualifier
RUN pip install -r requirements.txt
RUN rm -rf __pycache__
ADD . /app/monitoring
RUN mkdir -p /app/kml-input
RUN mkdir -p /app/flight-states
ENV PYTHONPATH /app
ARG version
ENV CODE_VERSION=$version
ARG qualifier_scd_version
ENV SCD_VERSION=$qualifier_scd_version
ENTRYPOINT []
