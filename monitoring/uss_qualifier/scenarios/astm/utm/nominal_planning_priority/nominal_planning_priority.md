# Nominal planning with priority test scenario

## Description

This test approximates normal strategic coordination where a user successfully
plans a flight whose operational intent is shared with other USSs, and where
another user takes priority over the first flight with an operational intent
with higher priority.

## Sequence

![Sequence diagram](sequence.png)

## Resources

### flight_intents

FlightIntentsResource that provides at least 2 flight intents.  The first flight intent will be planned normally and then the second flight will be planned on top of the first flight.  Therefore, the second flight must intersect the first flight, and the second flight must have higher priority than the first flight.

### flight_planners

FlightPlannersResource that provides exactly 2 flight planners (USSs).  The first flight planner will successfully plan the first flight.  The second flight planner successfully plan the second, higher-priority flight over the first one.

### dss

DSSInstanceResource that provides access to a DSS instance where flight creation/sharing can be verified.

## Setup test case

### Check for necessary capabilities test step

Both USSs are queried for their capabilities to ensure this test can proceed.

#### Valid responses check

If either USS does not respond appropriately to the endpoint queried to determine capability, this check will fail.

#### Support BasicStrategicConflictDetection check

This check will fail if the first flight planner does not support BasicStrategicConflictDetection.  If the second flight planner does not support HighPriorityFlights, this scenario will end normally at this point.

### Area clearing test step

Both USSs are requested to remove all flights from the area under test.

#### Area cleared successfully check

If either USS does not respond appropriately or fails to clear the area of operations, this check will fail.

## Plan first flight test case

### Inject flight intent test step

uss_qualifier indicates to the first flight planner a user intent to create the first flight.

#### Successful planning check

All flight intent data provided is correct and valid and free of conflict in space and time, therefore it should have been planned by the USS.  If the USS indicates a conflict, this check will fail.  If the USS indicates that the flight was rejected, this check will fail.  If the USS indicates that the injection attempt failed, this check will fail.

### Validate flight creation test step

TODO: uss_qualifier should verify that the flight actually planned is not too different from the flight request

### Validate flight sharing test step

This step verifies that the created flight is shared properly per ASTM F3548-21 by querying the DSS for flights in the area of the flight intent, and then retrieving the details from the USS if the operational intent reference is found.

#### DSS response check

If the DSS does not respond properly to the query that should yield the planned flight, this check will fail.

#### Operational intent shared correctly check

If a reference to the operational intent for the flight is not found in the DSS or the details cannot be retrieved from the USS, this check will fail and one of the requirements **ASTM F3548-21::USS0005** or **ASTM F3548-21::USS0105** were not met.

#### Correct operational intent details check

If the operational intent details reported by the USS do not match the user's flight intent, this check will fail.

## Plan priority flight test case

In this step, the second USS executes a user intent to plan a priority flight that conflicts with the first flight.

### Inject flight intent test step

#### Successful planning check

All flight intent data provided is correct and valid and free of conflict in space and time, therefore it should have been planned by the USS.  If the USS indicates a conflict, this check will fail since the intersecting flight is lower priority.  If the USS indicates that the flight was rejected, this check will fail.  If the USS indicates that the injection attempt failed, this check will fail.

## Activate priority flight test case

In this step, the second USS successfully executes a user intent to activate the priority flight.

TODO: Complete this test case

## Attempt to activate first flight test case

In this step, the first USS fails to activate the flight it previously created.

TODO: Complete this test case

## Cleanup

### Successful flight deletion check

Per **[scd.yaml::DeleteFlightSuccess](../../../../../interfaces/automated-testing/scd/scd.yaml)**, the deletion attempt of the previously-created flight should succeed for every flight planner under test.
