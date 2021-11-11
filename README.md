# USS to USS Communication and Synchronization [![Build Status](https://dev.azure.com/astm/dss/_apis/build/status/interuss.dss?branchName=master)](https://dev.azure.com/astm/dss/_build/latest?definitionId=2&branchName=master) [![GoDoc](https://godoc.org/github.com/interuss/dss?status.svg)](https://godoc.org/github.com/interuss/dss)

This repository contains the implementation of the Discovery and Synchronization Service (DSS) and a monitoring framework to test UAS Service Suppliers (USS). See the [InterUSS website](https://interuss.org) for background information.

**Standards and Regulations**

The DSS implementation and associated monitoring tools comply with the following standards and regulations:

- [ASTM F3411-19](https://www.astm.org/Standards/F3411.htm): Remote ID. [See OpenAPI interface](./interfaces/uastech/standards/remoteid)
- [ASTM WK63418](https://www.astm.org/DATABASE.CART/WORKITEMS/WK63418.htm): UAS Traffic Management (UTM) UAS Service Supplier (USS) Interoperability Specification. [See OpenAPI interface](./interfaces/astm-utm)

U-Space specific:

- [COMMISSION IMPLEMENTING REGULATION (EU) 2021/664](https://eur-lex.europa.eu/legal-content/EN/TXT/HTML/?uri=CELEX:32021R0664&from=EN#d1e32-178-1) *(work in progress)*

## Discovery and Synchronization Service (DSS)

The DSS is a simple and open service used by separate USSs, often in different organizations, to communicate information about UAS operations and coordinate with each other. This service is described in the ASTM remote ID standard. This flexible and distributed system is used to connect multiple USSs operating in the same general area to share information while protecting operator and consumer privacy. The system is focused on facilitating communication amongst actively operating USSs without details about UAS operations stored or processed in the DSS.

- [Introduction to the DSS implementation](./README_DSS.md)
- [Building and deploying a DSS instance](./build/README.md)
- [Conceptual background on the DSS and services](./concepts.md)
- [DSS implementation details](./implementation_details.md)

## Monitoring and UAS Service Suppliers (USS) testing

In addition to the DSS, this repository contains tools for USSs to test and validate their implementation of the services such as Remote ID (ASTM F3411) and Strategic Conflict Detection defined in ASTM WK63418, UAS Traffic Management (UTM) UAS Service Supplier (USS) Interoperability Specification.

- [Introduction to monitoring, conformance and interoperability testing](./monitoring/README.md)<br>Modules:
  - [USS Remote ID qualifier](./monitoring/rid_qualifier)
  <!-- - [USS SCD qualifier](./monitoring/scd_qualifier) -->
  - [DSSs interoperability tests](./monitoring/interoperability)
  - [DSS integration test: prober](./monitoring/prober)
  - [DSS load test](./monitoring/loadtest)
  - [Diagnostic tool to monitor DSS and USS interactions: tracer](./monitoring/tracer)

## Development Practices

<!-- - [Getting Started]() -->
<!-- - [Contribution Guidelines]() -->

- [Introduction to the repository](./introduction_to_repository.md)
- [Release process](./RELEASE.md)
<!-- - [Governance]() -->
