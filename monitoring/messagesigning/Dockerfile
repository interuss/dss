# start by pulling the python image
FROM python:3.8-alpine

RUN mkdir -p /app/monitoring/messagesigning
RUN mkdir -p /app/monitoring/monitorlib
COPY ./requirements.txt /app/monitoring/messagesigning/requirements.txt
COPY mockuss_well_known_public_key.py /app/monitoring/messagesigning/routes_well_known.py
# switch working directory
WORKDIR /app/monitoring/messagesigning

# install the dependencies and packages in the requirements file
RUN pip install -r requirements.txt

COPY . /app/monitoring/messagesigning

# configure the container to run in an executed manner
ENTRYPOINT [ "python" ]

CMD ["mockuss_well_known_public_key.py"]
