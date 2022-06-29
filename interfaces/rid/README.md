# ASTM F3411 Network remote identification

## F3411-19

[OpenAPI specification](v1/remoteid/augmented.yaml)

## F3411-xx (second version)

[OpenAPI specification](v2/remoteid/canonical.yaml)

## Mixing versions in a single ecosystem

If all USSs in an ecosystem use the v1 API, then everything is fine.  If all USSs in an ecosystem use the v2 API, then everything is fine.  If some USSs in an ecosystem use v1 while others use v2, there may be interoperability problems.  To avoid accidentally missing ISAs, this DSS implementation stores v1 and v2 ISAs alongside each other.  The URL field for v1 ISAs contains the `/flights` resource URL (e.g., `https://example.com/v1/uss/flights`), but this same URL field contains the base URL (e.g., `http://example.com/rid/v2`) for v2 ISAs.  This means a v1 USS may try to query `http://example.com/rid/v2` if reading a v2 USS's ISA, or a v2 USS may try to query `http://example.com/v1/uss/flights/uss/flights` if reading a v1 USS's ISA.  This issue is somewhat intentional because even though v1 and v2 both have a `/flights` endpoint, the communications protocol for these two endpoints is not compatible.  If v1 and v2 ISAs are going to co-exist in the same ecosystem, then every USS in that ecosystem must infer the USS-USS communications protocol based on the content of the URL field (`flights_url` and `identification_service_area_url` in v1 and `uss_base_url` in v2).

### v1 ISAs

If a USS in a mixed-version ecosystem reads a v1 ISA, it must communicate with the managing USS in the following ways:

| If the `flights_url` | Then reach `/flights` URL at | Using data exchange protocol |
| --- | --- | --- |
| Ends with "/flights" | `flights_url` without any changes | [F3411-19 (v1)](https://github.com/uastech/standards/blob/36e7ea23a010ff91053f82ac4f6a9bfc698503f9/remoteid/canonical.yaml#L325) |
| Does not end with "/flights" | `flights_url` + "[/uss/flights](https://github.com/uastech/standards/blob/ab6037442d97e868183ed60d35dab4954d9f15d0/remoteid/canonical.yaml#L338)" | [F3411-xx (v2)](https://github.com/uastech/standards/blob/ab6037442d97e868183ed60d35dab4954d9f15d0/remoteid/canonical.yaml#L338) |

### v1 SubscriberToNotify

If a USS in a mixed-version ecosystem makes a change to a v1 ISA, the response will [contain](https://github.com/uastech/standards/blob/36e7ea23a010ff91053f82ac4f6a9bfc698503f9/remoteid/canonical.yaml#L1263) a list of SubscriberToNotify.  The POST notification should be sent differently depending on the contents of the [`url` field](https://github.com/uastech/standards/blob/36e7ea23a010ff91053f82ac4f6a9bfc698503f9/remoteid/canonical.yaml#L1330):

| If the `url` | Then reach `/identification_service_areas` URL at | Using data exchange protocol |
| --- | --- | --- |
| Ends with "/identification_service_areas" | `url` + "/" + ISA ID | [F3411-19 (v1)](https://github.com/uastech/standards/blob/36e7ea23a010ff91053f82ac4f6a9bfc698503f9/remoteid/canonical.yaml#L767) |
| Does not end with "/identification_service_areas" | `url` + ["/uss/identification_service_areas/" + ISA ID](https://github.com/uastech/standards/blob/ab6037442d97e868183ed60d35dab4954d9f15d0/remoteid/canonical.yaml#L778) | [F3411-xx (v2)](https://github.com/uastech/standards/blob/ab6037442d97e868183ed60d35dab4954d9f15d0/remoteid/canonical.yaml#L822) |

### v2 ISAs

If a USS in a mixed-version ecosystem reads a v2 ISA, it must communicate with the managing USS in the following ways:

| If the `uss_base_url` | Then reach `/flights` URL at | Using data exchange protocol |
| --- | --- | --- |
| Ends with "/flights" | `uss_base_url` without any changes | [F3411-19 (v1)](https://github.com/uastech/standards/blob/36e7ea23a010ff91053f82ac4f6a9bfc698503f9/remoteid/canonical.yaml#L325) |
| Does not end with "/flights" | `uss_base_url` + "[/uss/flights](https://github.com/uastech/standards/blob/ab6037442d97e868183ed60d35dab4954d9f15d0/remoteid/canonical.yaml#L338)" | [F3411-xx (v2)](https://github.com/uastech/standards/blob/ab6037442d97e868183ed60d35dab4954d9f15d0/remoteid/canonical.yaml#L338) |

### v2 SubscriberToNotify

If a USS in a mixed-version ecosystem makes a change to a v2 ISA, the response will [contain](https://github.com/uastech/standards/blob/ab6037442d97e868183ed60d35dab4954d9f15d0/remoteid/canonical.yaml#L1475) a list of SubscriberToNotify.  The POST notification should be sent differently depending on the contents of the [`url` field](https://github.com/uastech/standards/blob/ab6037442d97e868183ed60d35dab4954d9f15d0/remoteid/canonical.yaml#L1545):

| If the `url` | Then reach `/identification_service_areas` URL at | Using data exchange protocol |
| --- | --- | --- |
| Ends with "/identification_service_areas" | `url` + "/" + ISA ID | [F3411-19 (v1)](https://github.com/uastech/standards/blob/36e7ea23a010ff91053f82ac4f6a9bfc698503f9/remoteid/canonical.yaml#L767) |
| Does not end with "/identification_service_areas" | `url` + ["/uss/identification_service_areas/" + ISA ID](https://github.com/uastech/standards/blob/ab6037442d97e868183ed60d35dab4954d9f15d0/remoteid/canonical.yaml#L778) | [F3411-xx (v2)](https://github.com/uastech/standards/blob/ab6037442d97e868183ed60d35dab4954d9f15d0/remoteid/canonical.yaml#L822) |
