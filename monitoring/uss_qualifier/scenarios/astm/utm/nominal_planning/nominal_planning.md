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

If either USS does not support BasicStrategicConflictDetection, then this check will fail.

### Area clearing test step

Both USSs are requested to remove all flights from the area under test.

#### Area cleared successfully check

If either USS does not respond appropriately or fails to clear the area of operations, this check will fail.

## Plan first flight test case

### Inject flight intent test step

uss_qualifier indicates to the first flight planner a user intent to create the first flight.

#### No conflict check

All flight intent data provided is correct and valid and free of conflict in space and time, therefore it should have been planned by the USS.  If the USS indicates a conflict, this check will fail.

#### Rejection check

All flight intent data provided was complete and correct with no airspace conflicts. The operational intent data should have been processed successfully and flight should have been planned.  If the USS indicates that the flight was rejected, this check will fail.

#### Failure check

All flight intent data provided was complete and correct. It should have been processed successfully, allowing the USS to reject or accept the flight.  If the USS indicates that the injection attempt failed, this check will fail.

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

## Attempt second flight test case

### Inject flight intent test step

#### Incorrectly planned check

The second flight intent conflicts with the first flight that was already planned.  If the USS successfully plans the flight, it means they failed to detect the conflict with the pre-existing flight, or else they modified the flight more than the user wanted.  Therefore, this check will fail if the second USS indicates success in creating the flight from the user flight intent.

#### Failure check

All flight intent data provided was complete and correct. It should have been processed successfully, allowing the USS to reject or accept the flight.  If the USS indicates that the injection attempt failed, this check will fail.

## Cleanup

### Successful flight deletion check

Per **[scd.yaml::DeleteFlightSuccess](../../../../../interfaces/automated-testing/scd/scd.yaml)**, the deletion attempt of the previously-created flight should succeed for every flight planner under test.
