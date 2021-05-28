FROM python:3.8
# Not -alpine because: https://stackoverflow.com/a/58028091/651139
# `docker build` should be run from `monitoring` (the parent folder of this folder)
RUN apt-get update && apt-get install openssl && apt-get install ca-certificates
RUN mkdir -p /app/monitoring/monitorlib
RUN mkdir /app/monitoring/mock_ridsp
COPY monitorlib/requirements.txt /app/monitoring/monitorlib/requirements.txt
COPY mock_ridsp/requirements.txt /app/monitoring/mock_ridsp/requirements.txt
WORKDIR /app/monitoring/mock_ridsp
RUN pip install -r requirements.txt
RUN rm -rf __pycache__
ADD . /app/monitoring
ENV PYTHONPATH /app
ARG version
ENV CODE_VERSION=$version

ENTRYPOINT []
