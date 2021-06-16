# RID Qualifier System Mock

## Overview
This folder containers a tool to simulate an entire remote ID ecosystem,
consisting of one or more USSs acting as Remote ID Service Providers, an
implicit DSS, and a USS acting as a Remote ID Display Provider, for the purpose
of testing `rid_qualifier`.  Once this system is running, `rid_qualifier` can
inject data into one or more virtual RID Service Providers supporting the
injection API and then observe the instantaneous state of the RID system via a
virtual RID Display Provider supporting the observation API.

## Execution

As long as the host system has Docker installed, simply run
[`run_locally.sh`](run_locally.sh) to start the rid_qualifier system mock
listening at localhost:8070.  Press CTRL/CMD-C to stop the system mock.

## Endpoints

### RID Service Providers
This mock supports any number of RID Service Providers, and they do not need to
be declared before use.  The RID Service Provider named `RIDSP` implements the
[injection API](../../../interfaces/automated-testing/rid/README.md) at
`http://hostname/sp/RIDSP`.  So, for instance, if an instance of this RID system
mock is accessible at localhost:8070, then data can be injected into the `uss1`
virtual RID Service Provider by making a PUT call to, e.g.,
http://localhost:8070/sp/uss1/tests/9a20678b-fad4-49e6-9009-b4891aa77cb7.

The behavior of each virtual RID Service Provider can be set PUTing JSON data to
the /sp/RIDSP/behavior endpoint according to the `ServiceProviderBehavior`
schema defined in [behavior.py].  Also see the sample requests in the Postman
collection (in Testing below).

### RID Display Providers
This mock supports any number of RID Display Providers, and they do not need to
be declared before use.  The RID Display Provider named `RIDDP` implements the
[observation API](../../../interfaces/automated-testing/rid) at
`http://hostname/dp/RIDDP`.  So, for instance, if an instance of this RID system
mock is accessible at localhost:8070, then the current flights visible to the
`uss1` virtual Display Provider can be queried by making a GET call to
http://localhost:8070/dp/uss1/display_data.

The behavior of each virtual RID Display Provider can be set PUTing JSON data to
the /dp/RIDDP/behavior endpoint according to the `DisplayProviderBehavior`
schema defined in [behavior.py].  Also see the sample requests in the Postman
collection (in Testing below).

## Debugging

To run the rid_qualifier system mock locally (without the Docker environment),
install the packages specified by [requirements.txt](requirements.txt) and then
run [`./debug_mock.py`](debug_mock.py), or the `gunicorn` command specified in
the `CMD` of the [`Dockerfile`](Dockerfile).

### Testing

To send test messages, open
[the provided collection](Postman_rid_qualifier_mock_test.json) in Postman and
send requests.  "Get status" indicates whether the system is running and
properly handling basic requests.  The "Injection" folder enables the injection
of a test with one simple ~5-minute flight with telemetry approximately once a
minute (and the deletion of that test).  After that test has been injected, the
"Observation" folder enables querying the apparent state of the virtual RID
system provided by the mock.  "Get display data (tight)" returns flights visible
in a small area.  "Get display data (loose)" intentionally requests a view that
is too large.  "Get display data (clusters)" requests a view larger than a
detailed view but small enough to be valid.
