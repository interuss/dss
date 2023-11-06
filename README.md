# USS to USS Discovery and Synchronization [![GoDoc](https://godoc.org/github.com/interuss/dss?status.svg)](https://godoc.org/github.com/interuss/dss)

<img src="assets/color_logo_transparent.png" width="200">

This repository contains InterUSS's implementation of the Discovery and Synchronization Service (DSS). See the [InterUSS website](https://interuss.org) for background information.

**The monitoring framework to test UAS Service Suppliers (USS) previously contained in this repository has moved to [a separate `monitoring` repository](https://github.com/interuss/monitoring).**

## Standards and Regulations

The DSS implementation targets compliance with the following standards and regulations:

- [ASTM F3411-19](https://www.astm.org/f3411-19.html) and [ASTM F3411-22](https://www.astm.org/f3411-22.html): Remote ID.
    - [F3411-19 OpenAPI interface](https://github.com/uastech/standards/releases/tag/astm_rid_1.0)
    - [F3411-22 OpenAPI interface](https://github.com/uastech/standards/releases/tag/astm_rid_api_2.1)
    - See [documentation](./interfaces/rid/README.md) before mixing versions in a single ecosystem.
- [ASTM F3548-21](https://www.astm.org/f3548-21.html): UAS Traffic Management (UTM) UAS
Service Supplier (USS) Interoperability Specification.
    - [F3548-22 OpenAPI interface](./interfaces/astm-utm)

U-Space specific:
- [COMMISSION IMPLEMENTING REGULATION (EU) 2021/664](https://eur-lex.europa.eu/legal-content/EN/TXT/HTML/?uri=CELEX:32021R0664&from=EN#d1e32-178-1)

## Discovery and Synchronization Service (DSS)

The DSS is a simple and open service used by separate USSs, often in different organizations, to communicate
information about UAS operations and coordinate with each other. This service is described in the ASTM remote
ID standard and ASTM USS interoperability standard. This flexible and distributed system is used to connect
multiple USSs operating in the same general area to share information while protecting operator and consumer
privacy. The system is focused on facilitating communication amongst actively operating USSs without details
about UAS operations stored in or processed by the DSS.

- [Introduction to the DSS implementation](./README_DSS.md)
- [Building and deploying a DSS instance](./build/README.md)
- [Conceptual background on the DSS and services](./concepts.md)
- [DSS implementation details](./implementation_details.md)

## Development Practices
- [Introduction to the repository](./introduction_to_repository.md)
- [Contributing](./CONTRIBUTING.md)
- [Release process](./RELEASE.md)
- [Governance](https://github.com/interuss/tsc)
