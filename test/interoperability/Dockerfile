FROM python:3.7-alpine
# `docker build` should be run from this folder
ADD . /app
WORKDIR /app
RUN pip install -r requirements.txt
RUN rm -rf __pycache__

ENTRYPOINT ["python", "interop.py"]
