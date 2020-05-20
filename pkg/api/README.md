All of these files are generated using a fork of
[https://github.com/nytimes/openapi2proto](https://github.com/nytimes/openapi2proto)
and [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway).

The proto-defined services are `rid` (remote ID), `scd` (strategic conflict
detection), and `aux` (auxiliary DSS endpoints). `rid` contains the standards-
compliant DSS proto service, `scd` contains the (draft) standards-compliant
DSS proto service, and `aux` adds some additional functionality. All are
generated from the yaml files present in interfaces/

All of the files can be generated from the Makefile at the root of this repo:

`make protos`

WARNING: DO NOT monkeypatch generated proto files by adding go libraries to the
generated packages. This can break future proto changes, as well as breaks
bazel based build systems.