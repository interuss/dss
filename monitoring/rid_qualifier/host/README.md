# RID Qualifier Host

`rid_qualifier host` is a development-level web server that provides endpoint to initialize the automated testing of the rid qualifier. 
This PR includes the initial setup for the web-server that can be run locally on your system via the (run_locally.sh)[run_locally.sh] script. 

## About

This project will create the following environment:

* redis Container: A redis server to handle background tasks.

* rq-worker Container: A container to run redis queue worker processes.

* rid-host Container: A flask application which accepts flight states' files, collect auth and config spec from user and executes RID Qualifier task. Once task finishes successfully, report.json file is available to download.

## Run via Docker

You should first create a new Python virtual environment to install the
required packages.

```bash
$ virtualenv ~/venv
$ source ~/venv/bin/activate
```

Change the root directory to dss/. Run the following command to start new containers:

```bash
(venv) $ ./monitoring/rid_qualifier/host/run_locally.sh
```

## Task Description

You can now test the application by uploading flight states json files. Input files are **required** to execute the task. When the flask application container is booted, a volume is
mounted from /tmp to /mnt/app/input-files to the containers where uploaded files are kept for processing.

A task may take few minutes to finish. A new task can not be launched until current task ends.

A sample response report.json can be downloaded by checking `Sample Report`.
