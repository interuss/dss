# Store NetRID flight data test scenario

## Overview

This scenario does not perform any checks, but instead writes NetRID flights data to files.  This may be useful to see or version-control the output of flight data generation, before it is used in a normal test suite.  Or, it may be useful to augment the standard report with exactly the flights data content being used in a normal test suite. 

## Resources

### flights_data

A [`FlightDataResource`](../../../resources/netrid/flight_data_resources.py) containing the flights data to write to files.

### storage_configuration

A [`FlightDataStorageResource`](../../../resources/netrid/flight_data_resources.py) describing where and how to write the flights data to files.

## Store flight data test case

### Store flight data test step

In this sole step for this entire test scenario, the provided flight data is written to file.
