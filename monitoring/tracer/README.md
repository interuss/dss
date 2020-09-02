# Diagnostic tool to monitor DSS and USS interactions

## Description
...

## Usage
From the [`root folder of this repo`](../..) folder:
```shell script
docker run --rm $(docker build -q -f monitoring/tracer/Dockerfile monitoring) \
    --auth=<SPEC> \
    --dss=https://example.com \
    --area=34.1234,-123.4567,34.4567,-123.1234
    --output-folder=logs
    --rid-isa-poll-interval=15
```

The auth SPEC defines how to obtain access tokens to access the DSS instances
and USSs in the network. See
[the auth spec documentation](../monitorlib/README.md#Auth_specs) for examples
and more information.
