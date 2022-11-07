# terraform-google-dss

This folder contains a terraform module which gathers resources used by all cloud providers.

It includes the automatic generation of the tanka configuration to deploy the Kubernetes resources
as well as the scripts required to generate the certificates and operate the cluster. 

See `examples/` for configuration examples.


## Configuration

See [variables.tf](./variables.tf) to configure the dss services.
