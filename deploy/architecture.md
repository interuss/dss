# Kubernetes deployment

## Introduction

See [introduction](../build/pooling.md#introduction)

## Target Architecture

The expected deployment of a DSS pool supporting a DSS Region consists of
multiple organizations hosting one DSS instance that is interoperable with
each other organization's DSS instance.  A DSS pool with three participating
organizations (USSs) will have an architecture similar to the diagram below.

_**Note** that the diagram shows 2 stateful sets per DSS instance.  Currently, the
files in this folder produce 3 stateful sets per DSS instance.  However, after
Issue #481 is resolved, this is expected to be reduced to 2 stateful sets._

![Pool architecture diagram](../assets/generated/pool_architecture.png)

### Terminology notes

See [teminology notes](../build/pooling.md#terminology-notes).

## Pooling

### Objective

See [Pooling Objective](../build/pooling.md#objective) and subsections.

### Additional requirements

See [Additional requirements](../build/pooling.md#additional-requirements).

### Survivability

See [survivability](../build/deploy/README.md#survivability).

### Sizing

See [sizing](../build/deploy/README.md#sizing).
