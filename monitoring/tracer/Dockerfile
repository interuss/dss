FROM python:3.7
# Not -alpine because: https://stackoverflow.com/a/58028091/651139
# `docker build` should be run from `monitoring` (the parent folder of this folder)
RUN mkdir -p /app/monitoring/monitorlib
RUN mkdir /app/monitoring/tracer
COPY monitorlib/requirements.txt /app/monitoring/monitorlib/requirements.txt
COPY tracer/requirements.txt /app/monitoring/tracer/requirements.txt
WORKDIR /app/monitoring/tracer
RUN pip install -r requirements.txt
RUN rm -rf __pycache__
ADD . /app/monitoring
ENV PYTHONPATH /app
ARG version
ENV CODE_VERSION=$version

ENTRYPOINT ["python"]
