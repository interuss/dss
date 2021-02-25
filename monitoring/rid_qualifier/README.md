# RID Qualifier Tests

This directory contains a series of tests for qualifying Network Remote ID compliance. It contains four things:

1. **Flight track dataset generator** (`flight_track_factory.py`): An API to generate multiple flight paths and patterns within a specified bounding box. You can specify a bounding box in any part of the world the generator will create a grid for flights for the bounds provided. In addition, circular flight paths will be generated within that grid. The flights will follow a circular path and these are generated in the form of latitude / longitude co-ordinates within the bounding box.

2. **Test payload generator** (`position_generator.py`): Once the flight tracks are generated, we simulate flights and "fly" them emitting positions to be fed to the test harness, a file is generated that can be submitted by the test harness to the API. You can specify how many minutes (in seconds) you want the flights to "loop" over the flight paths. Specific details metadata about the flights are generated and added in preparation for submission to the API.

3. **Test executor**: The test executor runs the test suite and generates flight paths and submits them to the test harness which in turn submits to the DSS.

4. **Remote ID Display Verifier**: Once the flights trajectories are sent to the DSS appropriate tests have to be carried on the Display provider side to check the compliance. We point `Flight Blender` a open-source Remote ID Display to test compliance.