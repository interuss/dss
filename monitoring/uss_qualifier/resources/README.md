## Overview

In order to execute a test scenario, it is normally necessary to provide configuration-specific resources that the scenario depends on.  This could include, for example:

* The set of NetRID Service Providers under test in a particular environment
* Flight data to be injected into the NetRID service providers, reflective of the jurisdiction in which the test is being conducted
* Performance evaluation criteria defining what additional latency or other factors may be introduced by the test apparatus

These things are provided to the test scenarios as resources via a named dependency injection framework described below.

## Resource types

A resource provider to a test scenario is an instance of a particular kind of resource.  This kind of resource is the "resource type", and resource types are implemented as subclasses of [the abstract `Resource` class](resource.py).  Therefore, a resource available to a test scenario will always be an instance of a subclass of `Resource`.

### Creating a new resource type

A new resource type is simply a Python subclass of [the abstract `Resource` class](resource.py).  It should be placed in the appropriate sub-module of the [`resources` module](.).  A requirement for any resource type is that the resource type must be constructable from 1) a specification for the resource and 2) any required dependencies (which must also be resources).

## Using resources in a test scenario

A test scenario requiring a resource should declare a named dependency on a resource of a particular type.  For example, the ASTM NetRID Nominal Behavior test scenario may declare that it has a `service_providers` dependency which must be satisfied with a `netrid.NetRIDServiceProviders` resource type.  The test scenario's configuration must specify the resource from the global resource pool (by specifying the globally-unique ID of the resource) that should be used to satisfy the test scenario's `service_providers` dependency.

## Declaring resources

Resources for a given test configuration are all declared in a single global resource collection, and then the resources generated from that collection are placed in a single global resource pool.  A resource from that pool may be referenced using its globally-unique (within the test configuration) identifier.  To declare a resource, a [`ResourceDeclaration`](resource.py) must be added to the test configuration's [`ResourceCollection`](resource.py).


2. A particular instance of test resource must have a globally unique name.
    * Example: `netrid.service_providers`, which would provide access to the RID injection API for each ASTM NetRID Service Provider under test.
    * Example: `netrid.flight_data.nominal_flights`, which would provide flight data for nominal flights which could be injected into Service Providers under test.
3. Every type of test resource must define a "resource specification", which is a serializable data type that fully defines how to create an instance of that resource type.
4. Every type of test resource must define how to create an instance of the test resource from an instance of the resource specification.
