# LoadTest tool

## Introduction
The LoadTest tool is based on [Locust](https://docs.locust.io/en/stable/index.html) which provides a UI for controlling the number of Users to spawn and make random requests. Currently its configured to make the request in the ratio 10 x Create ISA : 5 x Update ISA : 100 x Get ISA : 1 x Delete ISA. This means the User is 10 times likely to Create an ISA vs Deleting an ISA, and 10 times more likely to Get ISA vs Creating an ISA and so on. Subscription workflow is heavier on the Write side with the ratio of 100 x Create Sub : 50 x Update Sub : 20 x Get Sub : 5 x Delete Sub.

## Adjusting workload ratio
In each files every action has a weight declared in the `@task(n)` decorator. You can adjust the value of `n` to suite your needs

## Run locally without Docker
1. Go to the repository's root directory. We have to execute from root directory due to our directory structure choice.
1. Create virtrual environment `virtualenv --python python3 ./monitoring/loadtest/env`
1. Activate virtual environment `. ./monitoring/loadtest/env/bin/activate`
1. Install dependencies `pip install -r ./monitoring/loadtest/requirements.txt`
1. Set OAuth Spec with environment variable `AUTH_SPEC`. See [the auth spec documentation](../monitorlib/README.md#Auth_specs)
for the format of these values.  Omitting this step will result in Client Initialization failure.
1. You have 2 options of load testing the ISA or Subscription workflow

    a. For ISA run: `AUTH_SPEC="<auth spec>" locust -f ./monitoring/loadtest/locust_files/ISA.py -H <DSS Endpoint URL>`

    b. For Subscription run: `AUTH_SPEC="<auth spec>" locust -f ./monitoring/loadtest/locust_files/Sub.py -H <DSS Endpoint URL>`

## Running in a Container
Simply build the Docker container with the Dockerfile from the root directory. All the files are added into the container

1. From the [`monitoring`](..) folder of this repository, build the monitoring image with `docker image build . -t interuss/monitoring`
1. Run Docker container

    a. For ISA run: `docker run -e AUTH_SPEC="<auth spec>" -p 8089:8089 interuss/monitoring locust -f loadtest/locust_files/ISA.py`

    b. For Sub run: `docker run -e AUTH_SPEC="<auth spec>" -p 8089:8089 interuss/monitoring locust -f loadtest/locust_files/Sub.py`

1. If testing local DSS instance, be sure that the loadtest (monitoring) container has access to the DSS container

    a. For ISA run: `docker run -e AUTH_SPEC="DummyOAuth(http://host.docker.internal:8085/token,uss1)" --network="dss_sandbox_default" -p 8089:8089 interuss/monitoring locust -f loadtest/locust_files/ISA.py`

    b. For ISA run: `docker run -e AUTH_SPEC="DummyOAuth(http://host.docker.internal:8085/token,uss1)" --network="dss_sandbox_default" -p 8089:8089 interuss/monitoring locust -f loadtest/locust_files/Sub.py`

## Use
1. Navigate to http://localhost:8089
1. Start new test with number of Users to spawn and the rate to spawn them.
1. For the Host, provide the DSS Core Service endpoint used for testing. An example of such url is: http://dss_sandbox_local-dss-core-service_1:8082/v1/dss/ in case local environment is setup by [run_locally.sh](../../build/dev/run_locally.sh)
