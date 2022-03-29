# Multi node CockroachDB local cluster environment

This document describes deploying a local dss instance using an insecure multi-node CockroachDB cluster, and HAProxy load balancers to distribute client traffic.

# Setup

Multi-node local instance can be deployed by running [`./haproxy_local_setup.sh`](haproxy_local_setup.sh).

When this system is active (log messages stop being generated), the following
endpoints will be available:

* Dummy OAuth Server: http://localhost:8085/token
* DSS HTTP Gateway Server: http://localhost:8082/healthy
* CockroachDB web UI: http://localhost:8080



# Environment Details

Setup includes three cockroachdb nodes: roacha, roachb and roachc , each on a separate container, an HAProxy server: dss-crdb-cluster-for-testing, and the dss services. All the services are connecting to HAProxy server, which would distribute the traffic across the cluster. HAProxy, acting as a `load balancers` decouples client health from the health of a single CockroachDB node. In cases where a node fails, the load balancer redirects client traffic to available nodes.

Setting up HAProxy requires generating a configuration file by running `cockroach gen haproxy` on one of the cluster nodes. that is preset to work with the running cluster. Generated `haproxy.cfg` file is then mounted to HAProxy container via local machine's ~/Download/haproxy/ folder.

# Testing

In a different window, run [`./check_dss.sh`](check_dss.sh) to run a
demonstration RID query on the system.  The expected output is an empty list of
ISAs (no ISAs have been announced).
