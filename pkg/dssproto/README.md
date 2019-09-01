All of these files are generated using
[https://github.com/nytimes/openapi2proto](https://github.com/nytimes/openapi2proto)
and [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway) using
api.yaml present in the root level of this repository.

api.yaml is copied from
https://github.com/BenjaminPelletier/DiscoveryAndSynchronization/tree/c954debcf7511c918f5bc50904e7ab68e1c8dedd.

The upstream openapi2proto does not support Open API v3, so we use a forked
version from https://github.com/davidsansome/openapi2proto which supports enough
of it for our api.yaml.  This is done by a package replacement in go.mod.

To regenerate the files in this directory:

    cd InterUSS-Platform
    make pkg/dssproto/dss.pb.go pkg/dssproto/dss.pb.gw.go
