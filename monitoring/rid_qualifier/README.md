# RID Qualifier Tests

This directory contains a series of tests for qualifying tests for Network Remote ID. It contains four main things:

1. **Flight track dataset generator** (`flight_sim.py`): A API to generate flight paths and patterns within a bounding box. You can specify a bounding box in any part of the world the generator will create flight grid for the bounds and flight paths within that grid. The flights will follow a circular path and these are generated in the form of latitude / longitude within the bounding box.

2. **Position generator** (`flight_sim.py`): Once the flight tracks are generated, we simulate flights and "fly" them emitting position to be fed to the test harness, this is the `flight_sim` class. You can specify how many minutes (in seconds) you want the flights to "loop" over the flight paths. Specific details about the flights can be added as necessary. 

3. **Test executor**: The test executor runs the test suite and generates flight paths and submits them to the test harness which in turn submits to the DSS.

4. **Remote ID Display Verifier**: Once the flights trajectories are sent to the DSS appropriate tests have to be carried on the Display provider side to check the compliance. We point `Flight Blender` a open-source Remote ID Display to test compliance.