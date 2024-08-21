# Kubernetes deployment

## Introduction

See [introduction](../build/pooling.md#introduction)

## Definitions and terminology notes

This section defines some concepts and related references used in this documentation.

### DSS Region

A DSS Region is a region in which a single, unified airspace representation is
presented by one or more interoperable DSS instances, each instance typically
operated by a separate organization.  A specific environment (for example,
"production" or "staging") in a particular DSS Region is called a "pool".

### DSS Pool

A DSS Pool is a set of interoperable and interconnected DSS instances in a specific
DSS Region. Each instance is typically operated by a separate organization.

### DSS instance

A DSS instance is a single logical replica in a DSS pool hosted by a single
organization.

### Pooling

The process required by a DSS Instance to join a DSS Pool is referred to "Pooling"
in this documentation.

### CRDB cluster

CockroachDB (CRDB) establishes a distributed data store called a "cluster".
This cluster stores the DSS Airspace Representation (DAR) in multiple SQL
databases within that cluster.  This cluster is composed of many CRDB nodes,
potentially hosted by multiple organizations.
"CRDB cluster" is used to refer to this concept in this documentation.

### Kubernetes cluster

Kubernetes manages a set of services in a "cluster".  This is an entirely
different thing from the CRDB data store, and this type of cluster is what the
deployment instructions refer to.  A Kubernetes cluster contains one or more
node pools: collections of machines available to run jobs.  This node pool is an
entirely different thing from a DSS pool.
"Kubernetes cluster" is used to refer to this concept in this documentation.

## Architecture

See [architecture](../build/deploy/README.md#architecture)

## Pooling

### Objective

See [Pooling Objective](../build/pooling.md#objective) and subsections.

### Additional requirements

See [Additional requirements](../build/pooling.md#additional-requirements).

### Survivability

See [survivability](../build/deploy/README.md#survivability).

### Sizing

See [sizing](../build/deploy/README.md#sizing).
