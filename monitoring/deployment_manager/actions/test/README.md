## Contents

This folder contains deployment_manager actions that relate to tests not meant to exist in any real InterUSS deployment.

## hello_world

The actions in this module relate to a minimal nginx webserver.

### Deployment

To deploy a hello_world system, starting in the `deployment_manager` folder:

`python3 deployment_manager.py test/hello_world/deploy examples/test.json`

Once deployed, access the webserver via its ingress by running `minikube tunnel` and then visiting http://localhost, or access the webserver directly via its Service:

`minikube service webserver-service -n test`

To tear down a hello_world system, starting in the `deployment_manager` folder:

`python3 deployment_manager.py test/hello_world/destroy examples/test.json`

### Usage

To access the service: `minikube service webserver-service -n test`
