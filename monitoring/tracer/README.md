# Diagnostic tool to monitor DSS and USS interactions

## Description
This diagnostic tool monitors UTM traffic in a specified area.  This includes,
when requested, remote ID Identification Service Areas and Subscriptions, and
strategic deconfliction Operations, Constraints, and Subscriptions.  This tool
records data in a way not allowed in a standards-compliant production system, so
should not be run on a production system.

## Building the image
From the [`root folder of this repo`](../..) folder, first build the tracer
image:

```shell script
docker build \
    -f monitoring/tracer/Dockerfile \
    -t interuss/dss/tracer \
    --build-arg version=`scripts/git/commit.sh` \
    monitoring
```

## Polling mode
Polling mode periodically queries the DSS regarding the objects of interest and
notes when they appear, change, or disappear.  The primary advantage to this
mode is that it operates as a client only and does not require routing to
support an externally-accessible server.  One disadvantage is that fast changes
are not detected.  For instance, if an ISA was added and then deleted all within
a single polling period, this tool would not create an record of that ISA.

### Invocation
```shell script
docker run --rm -v `pwd`/logs:/logs interuss/dss/tracer \
    tracer_poll.py \
    --auth=<SPEC> \
    --dss=https://example.com \
    --area=34.1234,-123.4567,34.4567,-123.1234 \
    --output-folder=/logs \
    --rid-isa-poll-interval=15
```

The auth SPEC defines how to obtain access tokens to access the DSS instances
and USSs in the network. See
[the auth spec documentation](../monitorlib/README.md#Auth_specs) for examples
and more information.
