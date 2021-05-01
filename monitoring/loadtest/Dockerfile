FROM python:3.8
# Not -alpine because: https://stackoverflow.com/a/58028091/651139
# The context for this image should be the repo's root directory
RUN mkdir /monitoring
RUN mkdir /monitoring/prober
RUN mkdir /monitoring/monitorlib
RUN mkdir /monitoring/loadtest
COPY monitoring/loadtest /monitoring/loadtest
COPY monitoring/monitorlib /monitoring/monitorlib
COPY monitoring/prober /monitoring/prober
WORKDIR /monitoring/loadtest
RUN pip install -r requirements.txt
RUN rm -rf loadtest/__pycache__
ENV PYTHONPATH /
