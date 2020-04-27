All of these files are generated using
[https://github.com/nytimes/openapi2proto](https://github.com/nytimes/openapi2proto)
and [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway).

The 2 proto services are dss and aux. DSS represents the standards compliant DSS
proto service, while aux adds some additional functionality. Both are generated
from the yaml files present in interfaces/

All of the files can be generated from the Makefile:

`make protos`