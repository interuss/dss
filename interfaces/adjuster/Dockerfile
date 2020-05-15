FROM python:3.7-alpine
# `docker build` should be run from this folder
ADD ./requirements.txt /app/requirements.txt
RUN pip install -r /app/requirements.txt
RUN rm -rf __pycache__
ADD . /app
WORKDIR /app

ENTRYPOINT ["python", "adjust_openapi_yaml.py"]
