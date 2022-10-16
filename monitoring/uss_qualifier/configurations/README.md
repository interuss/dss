# uss_qualifier configurations

To execute a test run with uss_qualifier, a test configuration must be provided.  This configuration consists of the test suite to run, along with definitions for all resources needed by that test suite.  See [`TestConfiguration`](configuration.py) for the exact schema.

When referring to a configuration, three methods may be used:

* **Package-based**: refer to the .json file located in a subfolder of this `configurations` folder using the Python module style, omitting the extension of the file name.  For instance, `dev.local_test` would refer to [monitoring/uss_qualifier/configurations/dev/local_test.json](dev/local_test.json).
* **Local file**: when a configuration reference is prefixed with `file://`, it refers to a local file using the path syntax of the host operating system.
* **Web file**: when a configuration reference is prefixed with `http://` or `https://`, it refers to a file accessible at the specified URL.
