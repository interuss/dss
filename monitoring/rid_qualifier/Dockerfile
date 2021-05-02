FROM python:3.7
# Not -alpine because: https://stackoverflow.com/a/58028091/651139
# `docker build` should be run from `monitoring` (the parent folder of this folder)
RUN apt-get update && apt-get install openssl && apt-get install ca-certificates
RUN mkdir -p /app/monitoring/monitorlib
RUN mkdir /app/monitoring/rid_qualifer
COPY monitorlib/requirements.txt /app/monitoring/monitorlib/requirements.txt
COPY rid_qualifier/requirements.txt /app/monitoring/rid_qualifier/requirements.txt
WORKDIR /app/monitoring/rid_qualifer
RUN pip install -r requirements.txt
RUN rm -rf __pycache__
ADD . /app/monitoring
ENV PYTHONPATH /app
ARG version
ENV CODE_VERSION=$version
ENTRYPOINT ["python", "rid_qualifer.py"]