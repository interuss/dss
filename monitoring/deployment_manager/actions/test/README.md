## Contents

This folder contains deployment_manager actions that relate to tests not meant to exist in any real InterUSS deployment.

## hello_world

The actions in this module relate to a minimal nginx webserver.

### Deployment

To deploy a hello_world system, starting in the `deployment_manager` folder:

`python3 deployment_manager.py test/hello_world/deploy examples/test.json`

To tear down a hello_world system, starting in the `deployment_manager` folder:

`python3 deployment_manager.py test/hello_world/destroy examples/test.json`
