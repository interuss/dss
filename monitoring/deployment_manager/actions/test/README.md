## Contents

This folder contains deployment_manager actions that relate to tests not meant to exist in any real InterUSS deployment.

## hello_world

The actions in this module relate to a minimal webserver.

### Deployment

To deploy a hello_world system, starting in the `deployment_manager` folder:

`python3 deployment_manager.py test/hello_world/deploy examples/test.json`

To tear down a hello_world system, starting in the `deployment_manager` folder:

`python3 deployment_manager.py test/hello_world/destroy examples/test.json`

### Usage

Once deployed, the Service can be accessed directly with `minikube service webserver-service -n test`.  This will forward a port on the development machine to the Service port, and open a browser to localhost:THAT_PORT.

However, normal access to the system is usually provided via its ingress and is accomplished by running `minikube tunnel`.  While the tunnel is running:

* http://localhost or http://localhost:80 accesses the Ingress
* https://localhost (this sometimes works depending on browser) or https://localhost:443 (this should always work) access the Ingress
* http://localhost:8002 accesses the Service
* Port 8001 is not generally accessible outside the cluster, though the name of the pod can be discovered with `kubectl get pods -n test`, and then port 8001 from that pod can be forwarded to the development machine with, e.g., `kubectl -n test port-forward webserver-deployment-58f7879d9b-c2qqx 8001` (when this command is running, http://localhost:8001 in a browser will access the Pod directly).
