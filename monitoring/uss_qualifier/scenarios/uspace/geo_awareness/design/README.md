This folder contains information related to the design of the geo-awareness components of
the uss_qualifier automated testing suite.

## Architecture

Interactions between the uss_qualifier automated test driver and the USS under test are summarized in the diagram below.
The following steps are represented:
1. The test driver injects a reference to a Geo-Awareness dataset via the [Geo-Awareness test interface](../../../../../../interfaces/automated_testing/geo-awareness/v1/geo-awareness.yaml).
2. The USS under test loads the Geo-Awareness dataset. Meanwhile, the test driver polls the USS under test to track the loading progress until it is ready.
3. The test driver queries the USS under test via the test harness to validate that the USS successfully interprets the loaded dataset.

![Geo-Awareness architecture](./geo_awareness_architecture.png)

### Interfaces

The following sequence diagram outlines the interactions between the test driver and the test harness of the USS under test.  

![Geo-Awareness Test Driver interactions](./sequence.png)

### Test groups

* ED-269
