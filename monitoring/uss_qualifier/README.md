# USS Qualifier: Automated Testing

## Introduction

The USS Qualifier automates testing compliance to requirements and interoperability of multiple USS/USSPs.

## Architecture

Note: We are currently in the process of migrating the technical implementation of uss_qualifier to the architecture described here; the architecture today is different from this.

### Test suites

1. A test suite is a set of tests that establish compliance to the thing they're named after.
    * Example: Passing the "ASTM F3548-21" test suite should indicate the systems under test are compliant with ASTM F3548-21.
2. A test suite is composed of a list of {test suite|test scenario}.
    * Each element on the list is executed sequentially.

### Test scenarios

1. A test scenario is a logical, self-contained scenario designed to test specific sets of functionality of the systems under test.
    * The activities in a test scenario should read like a narrative of a simple play with a single plot.
2. A test scenario is composed of a list of test cases.
    * Each test case on the list is executed sequentially.
3. Test scenarios will usually be enumerated/identified/created by mapping a list of requirements onto a set of test scenarios (e.g., [NetRID](https://docs.google.com/spreadsheets/d/1YByckmK6hfMrGec53CxRM2BPvcgm6HQNoFxOrOEfrUQ/edit#gid=0), [Flight Authorisation](https://docs.google.com/spreadsheets/d/1IJkNS21Ps-2411LGhXBqWF7inQnPVeEA23dWjXpCR-M/edit#gid=0), [ED-269](https://docs.google.com/spreadsheets/d/1NIlRHtWzBXOyJ58pYimhDQDqsEyToTQRu2ma3AYXWEU/edit))
    * While unmapped requirements still exist: create new, simple test scenario that verifies a set of remaining unmapped requirements.
4. Most test scenarios will require test resources (like NetRID telemetry to inject, NetRID service providers under test, etc) customized to the ecosystem in which the tests are being performed; see [Test definitions](#test-definitions) below
    * A test scenario declares what kind of resource(s) it requires.

### Test cases

1. A test case is a single wholistic operation or action performed as part of a larger test scenario.
    * Test cases are like acts in the "play" of the test scenario they are a part of.
    * Test cases are typically the "gray headers” of the overview sequence diagrams.
2. A given test case belongs to exactly one test scenario.
3. A test case is composed of a list of test steps.
    * Each test step on the list is executed sequentially.

### Test steps

1. A test step is a single low-level task that must be performed in order to accomplish its associated test case.
   * Test steps are like scenes in the "play/act" of the test scenario/test case they are a part of.
2. A given test step belongs to exactly one test case.
3. A test step may have a list of checks associated with it.

### Checks

1. A check is the lowest-level thing automated testing does – it is a single pass/fail evaluation of a single criterion for a requirement.
2. A check evaluates information collected during the actions performed for a test step.
3. A given check belongs to exactly one test step.
4. Each check defines which requirement is not met if the check fails.

### Test configurations

1. Even though all the scenarios, cases, steps and checks are fully defined for a particular test suite, the scenarios require data customized for a particular ecosystem – this data is provided as "test resources" which are created from the specifications in a "test configuration".
2. A test configuration is associated with exactly one test suite, and contains descriptions for how to create each of the set of required test resources.
    * The resources required for a particular test definition depend on which test scenarios are included in the test suite.
3. One resource can be used by many different test scenarios.
4. One test scenario may use multiple resources.
5. One class of resources is resources that describe the systems under test and how to interact with them; e.g., "Display Providers under test".
    * This means that a complete test configuration can't be tracked in the InterUSS repository because it wouldn't make sense to list, e.g., Display Provider observation endpoint URLs in the SUSI qual-partners environment.
    * Partial test configurations, including RID telemetry to inject, operational intents to inject, etc, can be tracked in the InterUSS repository, but they could not be used without specifying the missing resources describing systems under test.

### [Test resources](resources/README.md)
