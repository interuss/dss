FROM python:3.8
# Not -alpine because: https://stackoverflow.com/a/58028091/651139
# `docker build` should be run from `monitoring` (the parent folder of this folder)
RUN mkdir -p /app/monitoring/interoperability
RUN mkdir /app/monitoring/monitorlib
COPY interoperability/requirements.txt /app/monitoring/interoperability/requirements.txt
COPY monitorlib/requirements.txt /app/monitoring/monitorlib/requirements.txt
WORKDIR /app/monitoring/interoperability
RUN pip install -r requirements.txt
RUN rm -rf __pycache__
ADD . /app/monitoring
ENV PYTHONPATH /app

ENTRYPOINT ["python", "interop.py"]
