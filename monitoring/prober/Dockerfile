FROM python:3.7
# Not -alpine because: https://stackoverflow.com/a/58028091/651139
# The context for this image should be monitoring/prober
RUN mkdir /app
COPY requirements.txt /app
WORKDIR /app
RUN pip install -r requirements.txt
RUN rm -rf __pycache__
ADD . /app

ENTRYPOINT ["pytest"]
