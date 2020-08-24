FROM python:3.7-alpine
# `docker build` should be run from `monitoring` (the parent folder of this folder)
RUN mkdir /app
RUN mkdir /app/interoperability
RUN mkdir /app/monitorlib
COPY interoperability/requirements.txt /app/interoperability/requirements.txt
COPY monitorlib/requirements.txt /app/monitorlib/requirements.txt
WORKDIR /app
RUN pip install -r interoperability/requirements.txt
RUN rm -rf __pycache__
COPY . /app

ENTRYPOINT ["python", "interop.py"]
