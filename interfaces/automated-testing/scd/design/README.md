This folder contains information related to the design of the SCD components of
the uss_qualifier automated testing suite.

## Architecture

Interactions between the uss_qualifier automated test driver and the system
under test (including USSs and a DSS) are summarized in the diagram below.  The
automated test driver attempts to inject test data, consisting of flights as
would be requested by a user, into USSs via the
[SCD test interface](../scd.yaml).  In addition to the responses to those
injection requests, the automated test driver also observes the consequences of
the injections using
[the standard DSS and USS SCD interfaces](../../../astm-utm/Protocol).

![Architecture diagram](strategic_coordination_architecture.png)

## Test groups

The tests performed by the automated test driver are grouped into the following
categories:

* [ASTM strategic coordination](astm-strategic-coordination)
* [U-space](u-space)
