# Plan valid flight test step

This page describes the content of a common test case where a valid user flight intent should be successfully planned by a flight planner.  See `inject_successful_flight_intent` in [test_steps.py](test_steps.py).

## Successful planning check

All flight intent data provided is correct and valid and free of conflict in space and time, therefore it should have been planned by the USS per **interuss.automated_testing.flight_planning.ExpectedBehavior**.  If the USS indicates a conflict, this check will fail.  If the USS indicates that the flight was rejected, this check will fail.  If the USS indicates that the injection attempt failed, this check will fail.
