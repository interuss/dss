FROM python:3.7-alpine
# The context for this image should be monitoring/prober
ADD . /app
WORKDIR /app
RUN pip install -r requirements.txt

ENTRYPOINT ["pytest", "."]
