`mock_ridsp` is a mock Remote ID Service Provider implementation.  It provides a
development-level web server that responds to requests to the USS endpoints
defined in the ASTM remote ID standard in a standards-compliant manner.

While responses for flights are currently no-ops, future development will add
the ability to inject flight data using the
[RID automated testing injection interface](../../interfaces/automated-testing/rid/README.md).
