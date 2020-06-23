FROM python:3.7-alpine
# The context for this image should be monitoring/prober
ADD . /app
WORKDIR /app
RUN pip install -r requirements.txt
RUN rm -rf __pycache__

ENTRYPOINT ["pytest"]
