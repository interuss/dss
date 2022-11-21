# USS Qualifier test suites

A test suite is a set of tests that establish compliance to the thing they're named after.  Example: Passing the "ASTM F3548-21" test suite should indicate the systems under test are compliant with ASTM F3548-21.

A test suite is composed of a list of {test suite|test scenario}; each element on the list is executed sequentially.

A test suite is defined with a YAML file following the [`TestSuiteDefinition` schema](definitions.py).
