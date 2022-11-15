# uss_qualifier configurations

## Usage

To execute a test run with uss_qualifier, a test configuration must be provided.  This configuration consists of the test suite to run, along with definitions for all resources needed by that test suite.  See [`TestConfiguration`](configuration.py) for the exact schema.

### Specifying

When referring to a configuration, three methods may be used; see [`FileReference` documentation](../fileio.py) for more details.

* **Package-based**: refer to a dictionary (*.json, *.yaml) file located in a subfolder of the `uss_qualifier` folder using the Python module style, omitting the extension of the file name.  For instance, `configurations.dev.local_test` would refer to [uss_qualifier/configurations/dev/local_test.json](dev/local_test.json).
* **Local file**: when a configuration reference is prefixed with `file://`, it refers to a local file using the path syntax of the host operating system.
* **Web file**: when a configuration reference is prefixed with `http://` or `https://`, it refers to a file accessible at the specified URL.

### Building

A valid configuration file must provide a single instance of the [`TestConfiguration` schema](configuration.py) in the format chosen (JSON or YAML), as indicated by the file extension (.json or .yaml).

#### Personalization

When designing personalized/custom configuration files for specific, non-standard systems, the configuration files should generally be stored in either [uss_qualifier/configurations/personal](personal), or in an external repository and provided via `file://` prefix or `http(s)://` prefix.

#### References

To reduce repetition in similar configurations, the configuration parser supports the inclusion of all or parts of other files by using a `$ref` notation similar to (but not the same as) OpenAPI.

When a `$ref` key is encountered, the keys and values of the referenced content are used to overwrite any keys at the same level of the `$ref`.  For instance:

_x.json_:
```json
{"a": 1, "$ref": "y.json", "b": 2}
```

_y.json_:
```json
{"b": 3, "c": 4}
```

Loading _x.json_ results in the object:

```json
{"a": 1, "b": 3, "c": 4}
```

To combine the contents from multiple `$ref` sources, use `allOf`.  For instance:

_q.json_:
```json
{"a": 1, "b": 2, "allOf": [{"$ref": "r.json"}, {"$ref": "s.json"}], "c": 3, "d":  4}
```

_r.json_:
```json
{"b": 5, "c": 6, "e":  7}
```

_s.json_:
```json
{"b": 8, "d": 9, "f": 10}
```

Loading _q.json_ results in the object:

```json
{"a": 1, "b": 8, "c": 6, "d": 9, "e": 7, "f": 10}
```

More details may be found in [`fileio.py`](../fileio.py).

## Design notes

1. Even though all the scenarios, cases, steps and checks are fully defined for a particular test suite, the scenarios require data customized for a particular ecosystem – this data is provided as "test resources" which are created from the specifications in a "test configuration".
2. A test configuration is associated with exactly one test suite, and contains descriptions for how to create each of the set of required test resources.
    * The resources required for a particular test definition depend on which test scenarios are included in the test suite.
3. One resource can be used by many different test scenarios.
4. One test scenario may use multiple resources.
5. One class of resources is resources that describe the systems under test and how to interact with them; e.g., "Display Providers under test".
    * This means that a complete test configuration can't be tracked in the InterUSS repository because it wouldn't make sense to list, e.g., Display Provider observation endpoint URLs in the SUSI qual-partners environment.
    * Partial test configurations, including RID telemetry to inject, operational intents to inject, etc, can be tracked in the InterUSS repository, but they could not be used without specifying the missing resources describing systems under test.
