# Flight planning action generators

This folder contains action generators related to USSs capable of flight planning.

## [FlightPlannerCombinations](planner_combinations.py) action generator

This action generator accepts a [FlightPlannersResource](../../resources/flight_planning/flight_planners.py) (note the plural) containing multiple USSs capable of flight planning and generates a specified action with combinations of flight planners from that resource provided as multiple [FlightPlannerResource](../../resources/flight_planning/flight_planners.py)s (note the singular) according to ResourceIDs specified as `roles`.  For instance, if a particular test scenario required two flight planners for the roles of `uss1` and `uss2`, a test designing might use a FlightPlannerCombinations to run that scenario multiple times according to a set of flight planner USSs to be tested.  If a `FlightPlannersResource` included {`ussA`, `ussB`, `ussC`} and the `FlightPlannerCombinations` action generator was configured to produce `ExampleTestScenario` actions from flight planner combinations for the roles `uss1` and `uss2`, then the scenarios below (or a subset, depending on configuration) would be produced as actions:

| `uss1` resource | `uss2` resource | Action                |
|-----------------|-----------------|-----------------------|
| `ussA`          | `ussA`          | `ExampleTestScenario` |
| `ussA`          | `ussB`          | `ExampleTestScenario` |
| `ussA`          | `ussC`          | `ExampleTestScenario` |
| `ussB`          | `ussA`          | `ExampleTestScenario` |
| `ussB`          | `ussB`          | `ExampleTestScenario` |
| `ussB`          | `ussC`          | `ExampleTestScenario` |
| `ussC`          | `ussA`          | `ExampleTestScenario` |
| `ussC`          | `ussB`          | `ExampleTestScenario` |
| `ussC`          | `ussC`          | `ExampleTestScenario` |

The usage intent for this action generator is to enable design of simple test scenarios with a small number of participants, but to automatically repeat that simple scenario with all applicable role assignment combinations given a list of flight planner USSs to test.
