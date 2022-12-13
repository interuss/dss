# Introduction to the Repository

This document aims to provide a introduction to the repository and it's structure to new contributors and developers.

## Repository structure

- The main codebase for the DSS is in `/pkg` and `/cmds`, the core organization and structure of the DSS is in these directories.
- The `/monitoring` folder contains tools to validate DSS deployments and has different directories detailing the tests.
- The `/interfaces` folder contains diagrams, API specifications of the standards and other test tools that come with the DSS. This folder contains references to the ASTM standard, diagrams about remote-id test suite etc.

### Introduction to the Monitoring toolset

The `monitoring` directory contains a set of folders containing different test suites to test different capabilities of the DSS during development and production use.

### Running tools locally

- When running tools in the monitoring toolset for local debugging outside of Docker, the [monitoring](https://github.com/interuss/dss/tree/master/monitoring) folder must be accessible from one of the entries in your `PYTHONPATH`.  To accomplish this, add the root repo folder to your `PYTHONPATH`.

- To run a DSS instance locally, see [this documentation](build/dev/standalone_instance.md).

### Building new monitoring tools

When building new monitoring tools, we recommend using Docker containers as a way to package and deploy / test them. Using Docker containers ensures that consistent runtime environments are created regardless of the OS / Architecture.

- When referring to services hosted on the host machine (whether in another Docker container or not) from a service in a Docker container, refer to the host machine from the service in a Docker container using `http://host.docker.internal` ([Windows documentation](https://docs.docker.com/docker-for-windows/networking/#use-cases-and-workarounds), [Mac documentation](https://docs.docker.com/docker-for-mac/networking/#use-cases-and-workarounds)). This often includes demonstration scripts like `run_locally.sh`

#### Prober

- The first and largest monitoring tool is the "prober" which a full integration test. The [prober documentation](monitoring/prober/README.md) describes how to run it. It uses the `pytest` framework to perform actions on a DSS instance and verify the results.
- With all of the monitoring tools, including prober, the deployment philosophy is that the monitoring folder is a Python package root so one can import, e.g., `monitoring.monitorlib.infrastructure`. The prober makes heavy use of the tools in side [/monitoring/monitorlib](monitoring/monitorlib/README.md) tools. Most of the time, reusable components are built in this library so other monitoring tools can use them.
- The way to set up access to the DSS in prober is to create special requests.Sessions which automatically perform their own authorization and have their own implicit USS identity. These sessions are created for dependency injection in [conftest.py](monitoring/prober/conftest.py).
- When we need requests to appear as if they are coming from two different USSs, we need to use two different Sessions. We can see that happening with `scd_session` and `scd_session2`.  The tests themselves are just functions prefixed with `test_` which is the standard way `pytest` manages test infrastructure.

#### Load test

- Another monitoring tool with a different binary is the [loadtest](monitoring/loadtest/README.md)
- This tool uses [Locust](https://locust.io) (a load testing tool/library) to send a bunch of requests to a DSS instance on an ongoing basis.  Again, you'll see that most of the complicated parts of interacting with the DSS instance are reused from `monitorlib`.

#### Tracer

- The [tracer](monitoring/tracer/README.md) is actually a set of two different entry points which act as a mock USS to examine and debug UTM traffic in a non-production deployment (*it cannot be used in a production deployment*).
- [tracer_poll](monitoring/tracer/tracer_poll.py) periodically polls the DSS for various object types (ISAs, Operations, Constraints) and if it finds anything new or different, it then queries the owning USS for the details of that object (when appropriate).
- [tracer_subscribe](monitoring/tracer/tracer_subscribe.py) creates Subscriptions in the DSS for various object types (ISAs, Operations, Constraints) and then listens for incoming push notifications from other USSs when there are changes to any of those monitored object types.  tracer_subscribe also exposes a small web interface to make its logs easily accessible (the ones from both tracer_subscribe and tracer_poll, if they are both running at the same time).
- [tracer_subscribe](monitoring/tracer/tracer_subscribe.py) also has the capability to make a single, one-off RID poll involving querying the DSS for ISAs, and then calling out to each applicable USS's /flights endpoint (and /flights/{id}/details, when appropriate) to get the current information.

#### USS Qualifier

The [uss_qualifier](monitoring/uss_qualifier/README.md) is a tool for testing / qualifying USS compliance with requirements including ASTM NetRID, flight planning, and more.
