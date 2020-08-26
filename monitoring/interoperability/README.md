# Tests for interoperability between multiple DSS instances

## Description
The test suite in this folder verifies that two DSS instances implementing
[the remote ID API](https://tiny.cc/dssapi_rid) in a shared DSS region
interoperate correctly.  This is generally accomplished by performing an
operation on one DSS instance and then verifying that the results are visible
in the other instance.  Neither of the two DSS instances need to be this
InterUSS Project implementation.

## Usage
From the [`root folder of this repo`](../..) folder:
```shell script
docker run --rm $(docker build -q -f monitoring/interoperability/Dockerfile monitoring) \
    --auth <SPEC> \
    --dss https://example.com/v1/dss \
    --dss https://example2.com/v1/dss
```

The auth SPEC defines how to obtain access tokens to access the DSS instances.
See [the prober documentation](../prober/README.md) for examples and more
information.
