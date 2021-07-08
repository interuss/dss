`rid_qualifier host` is a development-level web server that provides endpoint to initialize the automated testing of the rid qualifier. 
This PR includes the initial setup for the web-server that can be run locally on your system via the (run_locally.sh)[run_locally.sh] script. 

Upcoming feature PRs will include:
- A page with "Start test" button to execute the qualifier.
- Ability to run the job as a background task by adding it to the Redis Queue.
- Accept configuration object and auth spec as user input that can be passed as arguments to the code running as Redis worker.
- Invoke test_executor.main throguh the Redis worker.
- Update test_executor.main to return a Report object.
- Make the resulting report retrievable from the web interface.
