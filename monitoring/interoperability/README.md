# Tests for interoperability between multiple DSS instances

## Description
The test suite in this folder verifies that two DSS instances implementing
[the remote ID API](https://tiny.cc/dssapi_rid) in a shared DSS region
interoperate correctly.  This is generally accomplished by performing an
operation on one DSS instance and then verifying that the results are visible
in the other instance.  Neither of the two DSS instances need to be this
InterUSS Project implementation.

## Usage
From the [`root folder of this repo`](../..) folder:
```shell script
docker run --rm $(docker build -q -f monitoring/interoperability/Dockerfile monitoring) \
    --auth <SPEC> \
    https://example.com/v1/dss \
    https://example2.com/v1/dss
```

The auth SPEC defines how to obtain access tokens to access the DSS instances.
See [the auth spec documentation](../monitorlib/README.md#Auth_specs) for examples and more
information.

## Compliance with ASTM remote ID standard

### Standard-prescribed tests

A number of general tests are prescribed in A2.6.1 of the ASTM standard.  This
section expands on implementation details for those tests.  Sub-requirements are
added in bold in addition to the original standard text.

1. *PUT Identification Service Area:*  Tests must demonstrate that after an ISA
   is **(a)** created or **(b)** modified, it can **(c)** be retrieved from all
   DSS instances for the region with consistent results.  In addition, the end
   time for an ISA governs when the DSS automatically removes it from the DSS.
   Tests must demonstrate that **(d)** automatic removal of ISAs occurs on all
   DSS instances.
2. *DELETE Identification Service Area:*  Tests must demonstrate that an ISA can
   be **(a)** deleted on any DSS instance and **(b)** the deletion is reflected
   on all other DSS instances.   
3. *PUT Subscription:*  Tests must demonstrate that a subscription can be
   **(a)** created on any instance and notifications for the subscription are
   triggered when intersecting ISAs are **(b)** added or **(c)** modified to all
   other instances.  In addition, the end time for a subscription governs when
   the DSS automatically removes it from the DSS.  Tests must demonstrate that
   **(d)** automatic removal of subscriptions occurs on all DSS instances.  
4. *DELETE Subscription:*  Tests must demonstrate that that **(a)** a
   subscription can be deleted on any DSS instance and **(b)** the deletion is
   reflected on all other DSS instances.
5. *GET Subscription:*  Tests must demonstrate that a specific subscription can
   be retrieved from any DSS instance with consistent results.
6. *GET Subscriptions:*  Tests must demonstrate that the complete set of
   subscriptions in an area for a Net-RID Display Provider can be retrieved from
   any DSS instance with consistent results. 

### Test sequence

#### Legend

* *n*: Repeat for all N DSS instances; *n* denotes the current DSS index between
  1 and N
* *P*: Numeric DSS index of primary DSS under test.  The sequence below is
  intended to be repeated so that each DSS is the primary DSS under test for one
  iteration of the sequence.
* ISA[*id*]: Reference to Identification Service Area with a particular test id
  (test id index, not full UUID identifier).  Note that the same UUID should be
  used for ISA[i] throughout the sequence even though the logical ISA may be
  created and deleted multiple times.
* Subscription[*id*]: Reference to Subscription with a particular test *id*
  (test id index, not full UUID identifier).  Note that the same UUID should be
  used for Subscription[i] throughout the sequence even though the logical
  Subscription may be created and deleted multiple times.
* D: Number of seconds needed to process requests to all DSSs before the note to
  wait >D seconds from a particular time

#### Sequence

| Test | Action | Acceptance criteria | Qualitatively proves | A2.6.1
| --- | --- | --- | --- | ---
| S1 | USS1@DSS*P*: PUT ISA with no start time and end time 10 minutes from now | ISA[*P*] created with proper response | Can create ISA in primary DSS | 1a
| S2.n | USS2@DSS*n*: PUT Subscription with intersecting area, no start time | Subscription[*n*] created with proper response, service_areas includes ISA from S1.1 | Can create Subscription in primary DSS, ISA accessible from all non-primary DSSs | 3a, 1c
| S3.n | USS2@DSS*n*: GET Subscription[*P*] by ID | Subscription[*P*] returned with proper response | Can retrieve specific Subscription emplaced in primary DSS from all DSSs | 5
| S4.n | USS2@DSS*n*: GET Subscriptions using ISA[*P*]’s area | All Subscription[i] 1≤i≤n are returned in subscriptions with proper response | Can query all Subscriptions in area from all DSSs | 6
| S5 | USS1@DSS*P*: PUT ISA[*P*] setting end time to now + D seconds | ISA[*P*] modified with proper response, all Subscription[i] 1≤i≤n are returned in subscribers with proper response | Can modify ISA in primary DSS, ISA modification triggers subscription notification requests | 1b, 3c
| S6.n | USS2@DSS*P*: DELETE Subscription[*n*] | Subscription[*n*] deleted with proper response | Can delete Subscriptions in primary DSS | 4a
| S7.n | USS2@DSS*n*: GET Subscription[*n*] by ID | 404 with proper response | Subscription deletion from ID index was effective from primary DSS | 4b
| S8.n | USS2@DSS*n*: GET Subscriptions using ISA[*P*]’s area | No Subscription[i] 1≤i≤n returned with proper response | Subscription deletion from geographic index was effective from primary DSS | 4b
| S9.n | Wait >D seconds from S1.5 then USS2@DSS*n*: PUT Subscription with intersecting area, end time D seconds from now | Subscription[*n*] created with proper response, service_areas does not include ISA from S1.1/S1.5 | Expired ISA automatically removed, ISA modifications accessible from all non-primary DSSs | 1d, 1c
| S10 | USS1@DSS*P*: PUT ISA with no start time and end time 10 minutes from now | ISA[*P*] created with proper response, all Subscription[i] 1≤i≤n returned in subscribers with proper response | ISA creation triggers subscription notification requests | 3b
| S11 | USS1@DSS*P*: DELETE ISA[*P*] | ISA[*P*] deleted with proper response, all Subscription[i] 1≤i≤n returned in subscribers with proper response | ISA deletion triggers subscription notification requests | 3c, 2a
| S12 | Wait >D seconds from 1.9.n then USS1@DSS*P*: PUT ISA with no start time and end time 10 minutes from now | ISA[*P*] created with proper response, none of Subscription[i] 1≤i≤n returned in subscribers with proper response | Expired Subscriptions don’t trigger subscription notification requests | 3d
| S13.n | USS2@DSS*n*: GET Subscriptions using ISA[*P*]’s area | No Subscription[i] 1≤i≤n returned with proper response | Expired Subscription removed from geographic index on primary DSS | 3d
| S14.n | USS2@DSS*n*: GET Subscription[*n*] by ID | 404 with proper response | Expired Subscription removed from ID index on primary DSS | 3d
| S15 | USS1@DSS*P*: DELETE ISA[*P*] | ISA[*P*] deleted with proper response, none of Subscription[i] 1≤i≤n returned in subscribers with proper response | ISA deletion does not trigger subscription notification requests for expired Subscriptions | 3d
| S16.n | USS2@DSS*n*: PUT Subscription with intersecting area, no start time | Subscription[*n*] created with proper response, service_areas includes ISA from S1.12 | Deleted ISA removed from all DSSs | 2b
| S17.n | USS2@DSS*P*: DELETE Subscription[*n*] | Subscription[*n*] deleted with proper response | Nothing.  Action is a cleanup from test

#### Sequence diagram

![Sequence diagram for interoperability test](../../assets/generated/dss_interoperability_test.png)
