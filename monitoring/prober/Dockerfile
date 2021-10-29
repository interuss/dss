FROM python:3.8
# Not -alpine because: https://stackoverflow.com/a/58028091/651139
# The context for this image should be monitoring
RUN mkdir /monitoring
RUN mkdir /monitoring/prober
RUN mkdir /monitoring/monitorlib
COPY prober/requirements.txt /monitoring/prober
COPY monitorlib/requirements.txt /monitoring/monitorlib
WORKDIR /monitoring/prober
RUN pip install -r requirements.txt
RUN rm -rf __pycache__
ADD . /monitoring
ENV PYTHONPATH /monitoring

ENTRYPOINT ["pytest"]
