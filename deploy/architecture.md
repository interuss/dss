# Deployment of a DSS Pool

## Introduction

The DSS is designed to be deployed in a federated manner where multiple
organizations each host a DSS instance, and all of those instances interoperate.
Specifically, if a change is made on one DSS instance, that change may be read
from a different DSS instance.  A set of interoperable DSS instances is called a
"pool", and the purpose of this document is to describe how to form and maintain
a DSS pool.

It is expected that there will be exactly one production DSS pool for any given
DSS region, and that a DSS region will generally match aviation jurisdictional
boundaries (usually national boundaries).  A given DSS region (e.g.,
Switzerland) will likely have one pool for production operations, and an
additional pool for partner qualification and testing (per, e.g.,
F3411-19 A2.6.2).

## Architecture

See [architecture](../build/deploy/README.md#architecture)

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
