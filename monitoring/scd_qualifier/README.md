# SCD Qualifier Tests

This directory contains a series of tests for qualifying Strategic deconfliction (SCD) compliance per Annex IV of the U-Space regulation 2021/664 in the EU.

1. **Flight volume dataset generator** (`flight_volume_generator.py`): This module generates multiple flight paths the subsequent volumes given a bounding box. You can specify the number of paths that should be generated. The first path is called the `control` and the rest of the paths are generated using different options conflicting (or not) with this control. The last path is kept well clear to ensure no-intersection with the control. All the files are written to disk in the `test_definitions` directory. A additional JSON file is generated where the details of the intersection (or not) are written to check if the 4D submissions can be verified. The 4D volumes can then be submitted to the test harness.

2. **Test Executor** (`test_executor.py`): Currently not ready, this file executes the test by reading the test criteria and submitting the data to the harness and subsequently to the USSPs.