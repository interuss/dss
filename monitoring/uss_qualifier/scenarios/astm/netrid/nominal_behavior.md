# ASTM NetRID nominal behavior test scenario

## Overview

In this scenario, a single nominal flight is injected into each NetRID Service Provider (SP) under test.  Each of the injected flights is expected to be visible to all the observers at appropriate times and for appropriate requests. 

## Resources

### flights_data

A [`FlightDataResource`](../../../resources/netrid/flight_data.py) containing 1 nominal flight per SP under test.

### service_providers

A set of [`NetRIDServiceProviders`](../../../resources/netrid/service_providers.py) to be tested via the injection of RID flight data.  This scenario requires at least one SP under test.

### observers

A set of [`NetRIDObserversResource`](../../../resources/netrid/observers.py) to be tested via checking their observations of the NetRID system and comparing the observations against expectations.  An observer generally represents a "Display Application", in ASTM F3411 terminology.  This scenario requires at least one observer.

### evaluation_configuration

This [`EvaluationConfigurationResource`](../../../resources/netrid/observers.py) defines how to gauge success when observing the injected flights.

## Nominal flight test case

### Injection test step

In this step, uss_qualifier injects a single nominal flight into each SP under test, usually with a start time in the future.  Each SP is expected to queue the provided telemetry and later simulate that telemetry coming from an aircraft at the designated timestamps.

#### Successful injection check

Per **[injection.yaml::UpsertTestSuccess](../../../../../interfaces/automated-testing/rid/injection.yaml)**, the injection attempt of the valid flight should succeed for every NetRID Service Provider under test.

#### Valid flight check

Per **[injection.yaml::UpsertTestResult](../../../../../interfaces/automated-testing/rid/injection.yaml)**, the NetRID Service Provider under test should only make valid modifications to the injected flights.  This includes:
* A flight with the specified injection ID must be returned.

### Polling test step

In this step, all observers are periodically queried for the flights they observe.  Based on the known flights that were injected into the SPs in the previous step, these observations are checked against expected behavior/data.  Observation rectangles are chosen to encompass the known flights when possible.

#### Successful observation check

Per **[observation.yaml::ObservationSuccess](../../../../../interfaces/automated-testing/rid/observation.yaml)**, the call to each observer is expected to succeed since a valid view was provided by uss_qualifier.

#### Duplicate flights check

An assumption this test scenario currently makes is that the flight ID reported by the SP the flight was injected into is the same flight ID that each observer will report.  This is probably not a robust assumption and should be adjusted.

This check will fail if an observation contains two flights with the same ID.

#### Premature flight check

The timestamps of the injected telemetry usually start in the future.  If a flight with injected telemetry only in the future is observed prior to the timestamp of the first telemetry point, this check will fail because the SP does not satisfy **[injection.yaml::ExpectedBehavior](../../../../../interfaces/automated-testing/rid/injection.yaml)**.

#### Lingering flight check

**ASTM F3411-19::NET0260** and **ASTM F3411-22a::NET0260** require a SP to provide flights up to *NetMaxNearRealTimeDataPeriod* in the past, but an SP should preserve privacy and ensure relevancy by not sharing flights that are further in the past than this window.

#### Missing flight check

**ASTM F3411-19::NET0610** and **ASTM F3411-22a::NET0610** require that SPs make all UAS operations discoverable over the duration of the flight plus *NetMaxNearRealTimeDataPeriod*, so each injected flight should be observable during this time.  If one of the flights is not observed during its appropriate time period, this check will fail.

#### Area too large check

**ASTM F3411-19::NET0430** and **ASTM F3411-22a::NET0430** require that a NetRID Display Provider reject a request for a very large view area with a diagonal greater than *NetMaxDisplayAreaDiagonal*.  If such a large view is requested and a 413 error code is not received, then this check will fail.
