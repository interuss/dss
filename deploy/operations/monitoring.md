# Monitoring

## Prerequisites

Some of the tools from [the manual deployment documentation](../../build/README.md#prerequisites) are required to interact with monitoring services.

## Grafana / Prometheus

Note: this monitoring stack is only currently brought up when deploying [services](../README.md#deployment-layers) with [tanka](../services/tanka/README.md).

By default, an instance of Grafana and Prometheus are deployed along with the
core DSS services; this combination allows you to view (Grafana) CRDB metrics
(collected by Prometheus).  To view Grafana, first ensure that the appropriate
cluster context is selected (`kubectl config current-context`).  Then, run the
following command:

```shell script
kubectl get pod | grep grafana | awk '{print $1}' | xargs -I {} kubectl port-forward {} 3000
```

While that command is running, open a browser and navigate to
[http://localhost:3000](http://localhost:3000).  The default username is `admin`
with a default password of `admin`.  Click the magnifying glass on the left side
to select a dashboard to view.

## Prometheus Federation (Multi Cluster Monitoring)

The DSS can use [Prometheus](https://prometheus.io/docs/introduction/overview/) to
gather metrics from the binaries deployed with this project, by scraping
formatted metrics from an application's endpoint.
[Prometheus Federation](https://prometheus.io/docs/prometheus/latest/federation/)
enables you to easily monitor multiple clusters of the DSS that you operate,
unifying all the metrics into a single Prometheus instance where you can build
Grafana Dashboards for. Enabling Prometheus Federation is optional. To enable
you need to do 2 things:
1. Externally expose the Prometheus service of the DSS clusters.
2. Deploy a "Global Prometheus" instance to unify metrics.

### Externally Exposing Prometheus
You will need to change the values in the `prometheus` fields in your metadata tuples:
1. `expose_external` set to `true`
2. [Optional] Supply a static external IP Address to `IP`
3. [Highly Recommended] Supply whitelists of [IP Blocks in CIDR form](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing), leaving an empty list mean everyone can publicly access your metrics.
4. Then Run `tk apply ...` to deploy the changes on your DSS clusters.

### Deploy "Global Prometheus" instance
1. Follow guide to deploy Prometheus https://prometheus.io/docs/introduction/first_steps/
2. The scrape rules for this global instance will scrape other prometheus `/federate` endpoint and rather simple, please look at the [example configuration](https://prometheus.io/docs/prometheus/latest/federation/#configuring-federation).

## Health checks

This section describes various monitoring activities a USS may perform to verify various characteristics of their DSS instance and its pool.  In general, they rely on a DSS operator's monitoring infrastructure querying particular endpoints, evaluating the results of those queries, and producing alerts under certain conditions.  Not all checks listed below are fully implemented in the current InterUSS implementation.

One or more procedures below could be implemented into a single, more-accessible endpoint in monitoring middleware.

### /healthy check

#### Summary

Checks whether a DSS instance is responsive to HTTPS requests.

#### Procedure

For each expected DSS instance in the pool, query [`/healthy`](../../cmds/core-service/main.go)

#### Alert criteria

* Any query failed or returned a code other than 200

### Normal usage metrics

#### Summary

Checks whether normal calls to USS's DSS instance generally succeed.

#### Procedure

USS notifies its monitoring system whenever a normal ASTM-API call to its DSS instance fails due to an error indicating a failed service like timeout, 5xx, 405, 408, 418, 451, and possibly others.

#### Alert criteria

* Number of failures per time period crosses threshold

### DAR identity check

#### Summary

Checks whether a set of DSS instances indicate that they are using the same DSS Airspace Representation (DAR).

#### Procedure

For each expected DSS instance in the pool, query [`/aux/v1/pool`](../../interfaces/aux_) and collect `dar_id`

#### Alert criteria

* Any query failed
* Any collected `dar_id` value is different from any other collected `dar_id` value

### Per-USS heartbeat check

_Note: the implementation of this functionality is not yet complete._

#### Summary

Checks whether all DSS instance operators have recently verified their ability to synchronize data to another DSS instance operator.

#### Procedure

DSS instance operators agree to all configure their monitoring and alerting systems to execute this procedure, with an agreed-upon maximum time interval:

Assert a new heartbeat for the DSS operator's DSS instance via [`PUT /aux/v1/pool/dss_instances/heartbeat`](../../interfaces/aux_) which returns the list of `dss_instances` including each one's `most_recent_heartbeat`

#### Alert criteria

* `PUT` query fails
* Any expected DSS instance in the pool does not have an entry in `dss_instances`
* The current time is past any DSS instance's `next_heartbeat_expected_before`
* The difference between `next_heartbeat_expected_before` and `timestamp` is larger than the agreed-upon maximum time interval for any DSS instance

### Nonce exchange check

_Note: none of this functionality has been implemented yet._

#### Summary

Definitively checks whether pool data written into one DSS instance can be read from another DSS instance.

#### Implementation

This check would involve establishing the ability to read and write (client USS ID, DSS instance writer ID, nonce value) triplets in a database table describing pool information.

#### Procedure

1. For each expected DSS instance in the pool, write a nonce value
2. For each expected DSS instance in the pool, read all (DSS instance writer ID, nonce value) pairs written by the DSS instance operator's client USS ID

#### Alert criteria

* Any query failed
* The nonce value written to DSS instance i does not match the nonce value that DSS instance j reports was written to DSS instance i by the DSS instance operator

### DSS entity injection check

#### Summary

Actual DSS entities (subscriptions, operational intents) are manipulated in a geographically-isolated test area.

#### Procedure

Run [uss_qualifier](https://github.com/interuss/monitoring/tree/main/monitoring/uss_qualifier) with a suitable configuration.

The suitable configuration would cause DSS entities to be created, read, updated, and deleted within an isolated geographical test area, likely via a subset of the [dss all_tests](https://github.com/interuss/monitoring/blob/main/monitoring/uss_qualifier/suites/interuss/dss/all_tests.yaml) automated test suite with uss_qualifier possessing USS-level credentials.

#### Alert criteria

* Tested requirements artifact does not indicate Pass

### Database metrics check

#### Summary

Certain metrics exposed by the underlying database software are monitored.

#### Procedure

Each USS queries metrics of underlying database software ([CRDB](https://www.cockroachlabs.com/docs/stable/metrics), [YugabyteDB](https://docs.yugabyte.com/preview/launch-and-manage/monitor-and-alert/metrics/)) using their database node(s), including:

* Raft quorum availability

#### Alert criteria

* Any Raft quorum unavailability

## Failure detection capability

This section summarizes the preceding health checks and their ability to detect failures.

### Potential failures

_This list of failures and potential causes is not exhaustive in either respect._

1. DSS instance is not accepting incoming HTTPS requests
    1. Deployment not complete
    2. HTTP(S) ingress/certificates/routing/etc not configured correctly
    3. DNS not configured correctly
2. Database components of DSS instance are non-functional
    1. Database container not deployed correctly
    2. Database functionality failing
    3. Database software not behaving as expected
    4. Connectivity (e.g., username/password) between core-service and database not configured correctly
    5. System-range quorum of database nodes not met
    6. Trusted certificates for the pool not exchanged or configured correctly
3. USS initializes a stand-alone DSS instance or connects to a different pool rather than joining the intended pool
    1. Database initialization parameter not set properly during deployment + nodes to join omitted
    2. Nodes to join + trusted certificates specified incorrectly
4. USS shared the wrong base URL for their DSS instance with other pool participants
    1. I.e., USS deployed and uses fully-functional DSS instance at https://dss_q.uss.example.com connected to the DSS pool for environment Q, but indicates to other USSs that the DSS instance for environment Q is located at https://dss_r.uss.example.com (another fully-functional DSS instance connected to a different pool)
    2. _Note: the likelihood of this failure could be reduced to negligible if DSS base URLs were included in [#1140](https://github.com/interuss/dss/issues/1140)_
5. DSS instance can interact with the database, but cannot read from/write to any tables
    1. DSS instance operator executed InterUSS-unsupported manual commands directly to the database to change the access rights of database users used by DSS instances
6. DSS instance can read from and write to pool table, but cannot read from/write to SCD/RID tables
    1. DSS instance operator executed InterUSS-unsupported manual commands directly to the database to change the access rights of database users used by DSS instances on a per-table basis
    2. SCD/RID tables not initialized
    3. SCD/RID tables corrupt or not at appropriate schema version
7. The DSS instance connected to the pool is not the DSS instance the USS is using in the pool's environment
    1. USS specified the wrong DSS base URL in the rest of their system in the pool environment
        1. E.g., DSS instance at https://dss_x.uss.example.com is fully functional, connects to the DSS pool for environment X and is the base URL USS shares with other USSs, but the USS specifies https://dss_y.uss.example.com as the DSS instance for the rest of their system to use in environment X
    2. USS did not configure their system to use features (e.g., ASTM F3548-21 strategic coordination) requiring a DSS in the test of their system in the pool environment
8. DSS instance is working, but another part of the owning USS's system has failed
    1. USS deploys their DSS instance differently than/separately from the rest of their system, and the rest-of-system deployment failed while the DSS instance deployment is unaffected
    2. A component in the rest of the USS's system failed
9. Database software indicates success to the core-service client, but does not correctly synchronize data to other DSS instances
    1. There is a critical bug in the database software (this would seem to be a product problem rather than a configuration problem)
10. Aux API works but SCD/RID API does not work or is disabled
    1. DSS instance configuration does not enable SCD/RID APIs as needed
    2. SCD/RID endpoint routing does not work (though other routing does work)
11. Database nodes are unavailable such that quorum is not met for certain ranges
    1. Database node container(s) run out of disk space
    2. Database node container(s) are shut down due to resource shortage
    3. System maintenance conducted improperly (for instance, multiple USSs bring down nodes contributing to the same range for maintenance simultaneously)

### Check detection capabilities

<table>
    <tr>
        <th rowspan="2">Check</th>
        <th colspan="11">Failure</th>
    </tr>
    <tr>
        <th>1</th>
        <th>2</th>
        <th>3</th>
        <th>4</th>
        <th>5</th>
        <th>6</th>
        <th>7</th>
        <th>8</th>
        <th>9</th>
        <th>10</th>
        <th>11</th>
    </tr>
    <tr>
        <td>/healthy</td>
        <td>✅</td>
        <td>❌</td>
        <td>❌</td>
        <td>❌</td>
        <td>❌</td>
        <td>❌</td>
        <td>❌</td>
        <td>❌</td>
        <td>❌</td>
        <td>❌</td>
        <td>❌</td>
    </tr>
    <tr>
        <td>Normal usage metrics</td>
        <td>✅</td>
        <td>✅</td>
        <td>❌</td>
        <td>❌</td>
        <td>✅</td>
        <td>✅</td>
        <td>❌</td>
        <td>🔶</td>
        <td>❌</td>
        <td>✅</td>
        <td>🔶</td>
    </tr>
    <tr>
        <td>DAR identity</td>
        <td>✅</td>
        <td>✅</td>
        <td>✅</td>
        <td>✅</td>
        <td>❌</td>
        <td>❌</td>
        <td>❌</td>
        <td>❌</td>
        <td>❌</td>
        <td>❌</td>
        <td>🔶</td>
    </tr>
    <tr>
        <td>Per-USS heartbeat</td>
        <td>✅</td>
        <td>✅</td>
        <td>✅</td>
        <td>❌</td>
        <td>✅</td>
        <td>❌</td>
        <td>🔶</td>
        <td>🔶</td>
        <td>✅</td>
        <td>❌</td>
        <td>🔶</td>
    </tr>
    <tr>
        <td>Nonce exchange</td>
        <td>✅</td>
        <td>✅</td>
        <td>✅</td>
        <td>✅</td>
        <td>✅</td>
        <td>❌</td>
        <td>❌</td>
        <td>❌</td>
        <td>✅</td>
        <td>❌</td>
        <td>🔶</td>
    </tr>
    <tr>
        <td>DSS entity injection</td>
        <td>✅</td>
        <td>✅</td>
        <td>✅</td>
        <td>✅</td>
        <td>✅</td>
        <td>✅</td>
        <td>❌</td>
        <td>❌</td>
        <td>✅</td>
        <td>✅</td>
        <td>🔶</td>
    </tr>
    <tr>
        <td>Database metrics</td>
        <td>❌</td>
        <td>🔶</td>
        <td>❌</td>
        <td>❌</td>
        <td>❌</td>
        <td>❌</td>
        <td>❌</td>
        <td>❌</td>
        <td>🔶</td>
        <td>❌</td>
        <td>✅</td>
    </tr>
</table>
