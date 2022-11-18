# Nominal planning test scenario

## Overview

This test approximates normal strategic coordination where a user successfully
plans a flight whose operational intent is shared with other USSs, and where a
user cannot plan a flight because it would conflict with another operational
intent.

## Sequence

![Sequence diagram](sequence.png)

## Resources

### flight_intents

FlightIntentsResource that provides at least 2 flight intents.  The first flight intent will be used for the successfully-planned flight and the second flight will be used for the failed flight.  Therefore, the second flight must intersect the first flight.

### flight_planners

FlightPlannersResource that provides exactly 2 flight planners (USSs).  The first flight planner will successfully plan the first flight.  The second flight planner will unsuccessfully attempt to plan the second flight.

### dss

DSSInstanceResource that provides access to a DSS instance where flight creation/sharing can be verified.

## Setup test case

### Check for necessary capabilities test step

Both USSs are queried for their capabilities to ensure this test can proceed.

#### Valid responses check

If either USS does not respond appropriately to the endpoint queried to determine capability, this check will fail.

#### Support BasicStrategicConflictDetection check

If either USS does not support BasicStrategicConflictDetection, then this check will fail per **[astm.f3548.v21.GEN0310](../../../../requirements/astm/f3548/v21.md)** as the USS does not support the InterUSS implementation of that requirement.

### Area clearing test step

Both USSs are requested to remove all flights from the area under test.

#### Area cleared successfully check

**[interuss.automated_testing.flight_planning.ClearArea](../../../../requirements/interuss/automated_testing/flight_planning.md)**

## Plan first flight test case

### [Inject flight intent test step](../../../flight_planning/inject_successful_flight_intent.md)

The first flight intent should be successfully planned by the first flight planner.

### [Validate flight sharing test step](../validate_shared_operational_intent.md)

## Attempt second flight test case

### Inject flight intent test step

#### Incorrectly planned check

The second flight intent conflicts with the first flight that was already planned.  If the USS successfully plans the flight, it means they failed to detect the conflict with the pre-existing flight.  Therefore, this check will fail if the second USS indicates success in creating the flight from the user flight intent, per **[astm.f3548.v21.SCD0035](../../../../requirements/astm/f3548/v21.md)**.

#### Failure check

All flight intent data provided was complete and correct. It should have been processed successfully, allowing the USS to reject or accept the flight.  If the USS indicates that the injection attempt failed, this check will fail per **[interuss.automated_testing.flight_planning.ExpectedBehavior](../../../../requirements/interuss/automated_testing/flight_planning.md)**.

## Cleanup

### Successful flight deletion check

**[interuss.automated_testing.flight_planning.DeleteFlightSuccess](../../../../requirements/interuss/automated_testing/flight_planning.md)**
