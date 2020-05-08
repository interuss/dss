# Tests for interoperability between multiple DSS instances

## Description
The test suite in this folder verifies that two DSS instances implementing
[the remote ID API](https://tiny.cc/dssapi_rid) in a shared DSS region
interoperate correctly.  This is generally accomplished by performing an
operation on one DSS instance and then verifying that the results are visible
in the other instance.  Neither of the two DSS instances need to be this
InterUSS Project implementation.

## Usage
Run the test suite by installing the Python requirements in requirements.txt,
then running interop.py.  See below for a sandbox example.

## Sandbox example
...to be added...
