# LoadTest tool

The LoadTest tool is based on [Locust](https://docs.locust.io/en/stable/index.html) which provides a UI for controlling the number of Users to spawn and make random requests. Currently its configured to make the request in the ratio 10 x Create ISA : 5 x Update ISA : 100 x Get ISA : 1 x Delete ISA. This means the User is 10 times likely to Create an ISA vs Deleting an ISA, and 10 times more likely to Get ISA vs Creating an ISA and so on. Subscription workflow is heavier on the Write side with the ratio of 100 x Create Sub : 50 x Update Sub : 20 x Get Sub : 5 x Delete Sub.

# Adjusting workload ratio
In each files every action has a weight declared in the `@task(n)` decorator. You can adjust the value of `n` to suite your needs

# Run Locally without Docker
1. Go to the repository's root directory. We have to execute from root directory due to our directory structure choice.
1. Create vitrual environment `virtualenv --python python3 ./monitoring/loadtest/env`
1. Install dependencies `pip install -r ./monitoring/loadtest/requirements.txt`
1. Configure your OAuth Server Endpoint in `./monitoring/loadtest/locust_files/conf.py`

    a. Currently it only supports Dummy OAuth Server, if there is a request to support other Auth mechaisms, we can implement them.

1. Activate virtual environment `. ./monitoring/loadtest/env/bin/activate`
1. You have 2 options of load testing the ISA or Subscription workflow
    
    a. For ISA run: `locust -f ./monitoring/loadtest/locust_files/ISA.py -H <DSS Endpoint URL>`

    b. For Subscription run: `locust -f ./monitoring/loadtest/locust_files/Sub.py -H <DSS Endpoint URL>`

1. Navigate to http://localhost:8080
1. Start new test with number of Users to spawn and the rate to spawn them. 


# Running in a Container
Simply build the Docker container with the Dockerfile from the root directory. All the files are added into the container

1. Build command `docker build -f monitoring/loadtest/Dockerfile . -t loadtest`
1. Run Docker container

    a. For ISA run: `docker run loadtest locust -f locust_files/ISA.py`

    b. For Sub run: `docker run loadtest locust -f locust_files/Sub.py`