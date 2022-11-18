## Purpose

deployment_manager is an automation tool that will be capable of performing most of the steps a human would otherwise have to manually execute to:

* Deploy a DSS instance, both [locally](../../build/dev/standalone_instance.md) and [in production](../../build/README.md)
* [Troubleshoot a DSS instance](../../build/README.md)
* [Form or maintain a DSS pool](../../build/pooling.md)
* [Migrate database schemas](../../build/README.md#upgrading-database-schemas)
* [Upgrade CockroachDB nodes](https://www.cockroachlabs.com/docs/stable/upgrade-cockroach-version.html)
* Deploy a set of diagnostic tools ([tracer](../tracer), [uss_qualifier](../uss_qualifier), [mock_uss](../mock_uss))

deployment_manager actions all accept a declarative definition of the InterUSS deployment desired.  To make changes to a deployment, the administrator should edit the development definition and then execute the desired action.

## Usage

To use deployment_manager, first define an InterUSS deployment in a JSON file according to the [DeploymentSpec](systems/configuration.py) schema.  From a context with kubectl already configured to work with the target Kubernetes cluster, run [deployment_manager.py](deployment_manager.py) according to its help (`python3 deployment_manager.py --help`).

### Prerequisites

To prepare a system to run deployment manager, the following steps must be taken (commands relative to the root of this repository):

1. Install `monitorlib` requirements (`pip3 install -r monitoring/monitorlib/requirements.txt`)
1. Install `deployment_manager` requirements (`pip3 install -r monitoring/deployment_manager/requirements.txt`)
1. Make `monitoring` accessible as a Python module source (`export PYTHONPATH=$(pwd)`)
1. See the syntax for running deployment_manager (`python3 monitoring/deployment_manager/deployment_manager.py --help`)

## Local development

deployment_manager is designed to interact with deployments managed with Kubernetes.  To host a Kubernetes deployment on your local development machine, the use of a minikube cluster is recommended.

### Install minikube

* Follow [installation instructions](https://minikube.sigs.k8s.io/docs/start/) up to, and including, "Start your cluster"
* [Enable ingress](https://kubernetes.io/docs/tasks/access-application-cluster/ingress-minikube/) at 127.0.0.1 with `minikube addons enable ingress`

### minikube usage

If minikube is already installed, then start a session with `minikube start`.

If kubectl is not currently configured to control minikube: `kubectl config use-context minikube`

View the minikube dashboard with `minikube dashboard`.

### Hello world

The actions in [test/hello_world](actions/test/hello_world.py) demonstrate manipulation of [a simple system](actions/test/README.md) using deployment_manager; see [the README](actions/test/README.md) for actions that can be performed on this system.
