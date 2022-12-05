# Multi node CockroachDB local cluster environment

This document describes deploying a local dss instance using an insecure multi-node CockroachDB cluster and HAProxy load balancers to distribute client traffic.


# Setup

Multi-node local instance can be deployed by running [`./build/dev/haproxy_local_setup.sh`](haproxy_local_setup.sh).

Once the setup is complete, the following endpoints will be available:

* Dummy OAuth Server: http://localhost:8085/token
* DSS Core Service: http://localhost:8082/healthy
* CockroachDB web UI: http://localhost:8080

Run [`./check_dss.sh`](check_dss.sh) to ensure environment is up. The expected output is an empty list of ISAs (no ISAs have been announced).

Run `./build/dev/haproxy_local_setup.sh down` to stop the system.

# Environment Details

Setup includes three cockroachdb nodes: roacha, roachb and roachc , each on a separate container, an HAProxy server: dss-crdb-cluster-for-testing, and the dss services. All the services are connecting to HAProxy server, which would distribute the traffic across the cluster. HAProxy, acting as a `load balancers` decouples client health from the health of a single CockroachDB node. In cases where a node fails, the load balancer redirects client traffic to available nodes.

Setting up HAProxy requires generating a configuration file by running `cockroach gen haproxy` on one of the cluster nodes. that is preset to work with the running cluster. Generated `haproxy.cfg` file is then mounted to HAProxy container via local machine's ~/Download/haproxy/ folder.


# Testing

To test the DB connection through HAProxy, run [`./build/dev/check_scd_write.sh`](check_scd_write.sh) to perform some write operations on the database.


Stop one of the cluster nodes by running:

    ```docker container stop roacha```

DSS services should still be up and running. Test it by running read operations:

    ```./build/dev/check_scd_read.sh```

Run following to bring up a new node in the cluster environment:

    ```docker container run -d --name roachc \
        --hostname roachc -p 8087:8087  \
        --network dss_sandbox_default \
        cockroachdb/cockroach:v21.2.3 start --insecure --join=roacha,roachb
    ```

Repeat calling `check_*.sh` scripts to test the cluster health.
