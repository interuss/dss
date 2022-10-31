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

### [Test scenarios](scenarios/README.md)

### Test cases

1. A test case is a single wholistic operation or action performed as part of a larger test scenario.
    * Test cases are like acts in the "play" of the test scenario they are a part of.
    * Test cases are typically the "gray headers” of the overview sequence diagrams.
2. A given test case belongs to exactly one test scenario.
3. A test case is composed of a list of test steps.
    * Each test step on the list is executed sequentially.

### Test steps

1. A test step is a single task that must be performed in order to accomplish its associated test case.
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

