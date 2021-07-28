`rid_qualifier host` is a development-level web server that provides endpoint to initialize the automated testing of the rid qualifier. 
This PR includes the initial setup for the web-server that can be run locally on your system via the (run_locally.sh)[run_locally.sh] script. 

**Prerequisite**: Local DSS instance should be started by running (monitoring/rid_qualifier/mock/run_locally.sh)[monitoring/rid_qualifier/mock/run_locally.sh] script.

**Task Description**: A `Start Test` button on the home page allows user to provide config and auth spec to run the RID qualifier Host and a submit button to execute the test_executor task. On successfully completing the job, user gets a link to download the resulting `report.json` file.

A task may take few minutes to finish. A new task can not be launched until current task ends.

A sample response report.json can be downloaded by checking `Sample Report`.
