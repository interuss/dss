# SCD Qualifier Tests
This directory contains a series of tests for qualifying Strategic deconfliction (SCD) compliance.


1. **Flight volume dataset generator** (`flight_volume_generator.py`): This module generates multiple flight paths and volumes given a bounding box. You can specify the number of paths that should be generated. The first path is called the `control` and the rest of the paths are generated using different options intersecting with this control. The last path is kept well clear to ensure no-intersection with the control. All the files are written to disk in the `test_definitions` directory. The 4D volumes can then be submitted to the 
