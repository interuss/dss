# DSS pool monitoring

This page describes various monitoring activities a USS may perform to verify various characteristics of their DSS instance and its pool.

## Checks

### /healthy check

#### Summary

Checks whether a DSS instance is responsive to HTTPS requests.

#### Procedure

For each expected DSS instance in the pool, query `/healthy`

#### Alert criteria

* Any query failed or returned a code other than 200

### DSS Airspace Representation identity check

#### Summary

Checks whether a set of DSS instances indicate that they are using the same DSS Airspace Representation.

#### Procedure

For each expected DSS instance in the pool, query `/aux/v1/pool` and collect `dar_id`

#### Alert criteria

* Any query failed
* Any collected `dar_id` value is different from any other collected `dar_id` value

### Per-USS heartbeat check

_Note: the implementation of this functionality is not yet complete._

#### Summary

Checks whether all DSS instance operators have recently verified their ability to synchronize data to another DSS instance operator.

#### Procedure

DSS instance operators agree to all configure their monitoring and alerting systems to execute this procedure, with an agreed-upon maximum time interval:

Assert a new heartbeat for the DSS operator's DSS instance via `PUT /aux/v1/pool/dss_instances/heartbeat` which returns the list of `dss_instances` including each one's `most_recent_heartbeat`

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
2. For each expected DSS instance in the pool, read all (DSS instance writer ID, nonce value) pairs from the DSS instance operator's client USS ID

#### Alert criteria

* Any query failed
* The nonce value written to DSS instance i does not match the nonce value that DSS instance j reports was written to DSS instance i by the DSS instance operator

### DSS entity injection check

#### Summary

Actual DSS entities (subscriptions, operational intents) are manipulated in a geographically-isolated test area.

#### Procedure

Run uss_qualifier with a suitable configuration.

The suitable configuration would cause DSS entities to be created, read, updated, and deleted within an isolated geographical test area, likely via a subset of the dss_probing automated test with uss_qualifier possessing USS-level credentials.

#### Alert criteria

* Tested requirements artifact does not indicate Pass

## Failure detection capability

This section summarizes the preceding checks and their ability to detect failures.

### Potential failures

1. DSS instance is not accepting incoming HTTPS requests
2. Database components of DSS instance are non-functional
3. USS initializes a stand-alone DSS instance (or connected to a different pool) rather than joining the intended pool
4. DSS instance can interact with the database, but cannot read from/write to any tables
5. The DSS instance connected to the pool is not the DSS instance the USS is using in the pool's environment
6. DSS instance is working, but another part of the owning USS's system has failed
7. Database software indicates success to the core-service client, but does not correctly synchronize data to other DSS instances
8. DSS instance can read from and write to pool table, but cannot read from/write to SCD/RID tables
9. Aux API works but SCD/RID API does not work or is disabled

### Check detection capabilities

<table>
    <tr>
        <th rowspan="2">Check</th>
        <th colspan="9">Failure</th>
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
    </tr>
    <tr>
        <td>DSS Airspace Representation identity</td>
        <td>✅</td>
        <td>✅</td>
        <td>✅</td>
        <td>❌</td>
        <td>❌</td>
        <td>❌</td>
        <td>❌</td>
        <td>❌</td>
        <td>❌</td>
    </tr>
    <tr>
        <td>Per-USS heartbeat</td>
        <td>✅</td>
        <td>✅</td>
        <td>✅</td>
        <td>✅</td>
        <td>🔶</td>
        <td>🔶</td>
        <td>❌</td>
        <td>❌</td>
        <td>❌</td>
    </tr>
    <tr>
        <td>Nonce exchange</td>
        <td>✅</td>
        <td>✅</td>
        <td>✅</td>
        <td>✅</td>
        <td>❌</td>
        <td>❌</td>
        <td>✅</td>
        <td>❌</td>
        <td>❌</td>
    </tr>
    <tr>
        <td>DSS entity injection</td>
        <td>✅</td>
        <td>✅</td>
        <td>✅</td>
        <td>✅</td>
        <td>❌</td>
        <td>❌</td>
        <td>✅</td>
        <td>✅</td>
        <td>✅</td>
    </tr>
</table>
