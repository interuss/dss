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

This check will fail if the first flight planner does not support BasicStrategicConflictDetection per **astm.f3548.v21.GEN0310** as the USS does not support the InterUSS implementation of that requirement.  If the second flight planner does not support HighPriorityFlights, this scenario will end normally at this point.

### Area clearing test step

Both USSs are requested to remove all flights from the area under test.

#### Area cleared successfully check

**interuss.automated_testing.flight_planning.ClearArea**

## Plan first flight test case

### [Inject flight intent test step](../../../flight_planning/inject_successful_flight_intent.md)

The first flight intent should be successfully planned by the first flight planner.

### [Validate flight sharing test step](../validate_shared_operational_intent.md)

## Plan priority flight test case

In this step, the second USS executes a user intent to plan a priority flight that conflicts with the first flight.

### [Inject flight intent test step](../../../flight_planning/inject_successful_flight_intent.md)

The first flight intent should be successfully planned by the first flight planner.

### [Validate flight sharing test step](../validate_shared_operational_intent.md)

## Activate priority flight test case

In this step, the second USS successfully executes a user intent to activate the priority flight.

TODO: Complete this test case

## Attempt to activate first flight test case

In this step, the first USS fails to activate the flight it previously created.

TODO: Complete this test case

**astm.f3548.v21.SCD0015**

## Cleanup

### Successful flight deletion check

**interuss.automated_testing.flight_planning.DeleteFlightSuccess**
