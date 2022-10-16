# Test scenarios

## Definition

A test scenario is a logical, self-contained scenario designed to test specific sets of functionality of the systems under test.  The activities in a test scenario should read like a narrative of a simple play with a single plot.

## Structure

A test scenario is separated into of a list of test cases.  Each test case on the list is executed sequentially.  A test case is like an act in a play.

## Creation

Test scenarios will usually be enumerated/identified/created by mapping a list of requirements onto a set of test scenarios (e.g., [NetRID](https://docs.google.com/spreadsheets/d/1YByckmK6hfMrGec53CxRM2BPvcgm6HQNoFxOrOEfrUQ/edit#gid=0), [Flight Authorisation](https://docs.google.com/spreadsheets/d/1IJkNS21Ps-2411LGhXBqWF7inQnPVeEA23dWjXpCR-M/edit#gid=0), [ED-269](https://docs.google.com/spreadsheets/d/1NIlRHtWzBXOyJ58pYimhDQDqsEyToTQRu2ma3AYXWEU/edit)).  To form a complete set of scenarios to cover a set of requirements:

    While unmapped requirements still exist:
        Create new, simple test scenario that verifies a set of remaining unmapped requirements.

## Resources

Most test scenarios will require [test resources](../resources/README.md) (like NetRID telemetry to inject, NetRID service providers under test, etc) usually customized to the ecosystem in which the tests are being performed.  A test scenario declares what kind of resource(s) it requires, and a test suite identifies which available resources should be used to fulfill each test scenario's needs.

## Documentation

Traceability between requirements and test activities is of the utmost importance in automated testing.  As such, every test scenario must be documented, and that documentation must follow a precise format.  Conformance to this format is [checked by an automated test](../scripts/validate_test_definitions.sh) before changes to test scenarios or their documentation can be submitted to this repository.

Documentation requirements include:

### Documentation location

The documentation must be located in a .md file with the same name as the Python file that defines the `TestScenario`.  For instance, if a `NominalBehavior` class inherited from `TestScenario` and was defined in nominal_behavior.py, then documentation for `NominalBehavior` would be expected in nominal_behavior.md located in the same folder as nominal_behavior.py.

### Scenario name

The first line of the documentation file must be a top-level section header with the name of the test scenario ending in " test scenario".  Example: `# ASTM NetRID nominal behavior test scenario`

### Resources

A main section in the documentation must be named "Resources" (example: `## Resources`).  This section must have a subsection for each resource required by the test scenario, and each of these sections must be named according to the parameter in the `TestScenario` subclass's constructor for that resource.  For example, if a test scenario were defined as:

```python
class NominalBehavior(TestScenario):
    def __init__(self, flights_data: FlightDataResource,
                 service_providers: NetRIDServiceProviders:
        ...
```

...then the Resources section (`# Resources`) of the documentation would be expected to have two subsections: one for `flights_data` (`## flights_data`) and one for `service_providers` (`## service_providers`).  These sections should generally explain the purpose, use, expectations, and/or requirements for the resources.

### Test cases

A scenario must document at least one test case (otherwise the scenario is doing nothing).  Each test case must be documented via a main section in the documentation named with a " test case" suffix (example: `## Nominal flight test case`).

### Test steps

Each test case in the documentation must document at least one test step (otherwise nothing is happening in the test case).  Each test step must be documented via a subsection of the parent test case named with a " test step" suffix (example: `### Injection test step`).

### Test checks

Each check a test step performs that may result in a finding/issue must be documented via a subsection of the parent test step, named with a " check" suffix (example: `#### Successful injection check`).

A check should document the requirement(s) violated if the check fails.  Requirements are identified by putting a strong emphasis/bold style around the requirement ID (example: `**ASTM F3411-19::NET0420**`).
