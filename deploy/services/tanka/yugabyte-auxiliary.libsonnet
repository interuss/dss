# TODO: This file is not implemented yet
local base = import 'base.libsonnet';
local volumes = import 'volumes.libsonnet';


local googleYugabyteLB(metadata, name, ip) = base.Service(metadata, name) {
  port:: metadata.yugabyte.grpc_port,
  app:: 'yugabyte',
  spec+: {
    type: 'LoadBalancer',
    loadBalancerIP: ip,
  },
};

local awsYugabyteLB(metadata, name, ip) = base.AWSLoadBalancer(metadata, name, [ip], metadata.subnet) {
  port:: metadata.yugabyte.grpc_port,
  app:: 'yugabyte',
};

local minikubeYugabyteLB(metadata, name, ip) = base.Service(metadata, name) {
  port:: metadata.yugabyte.grpc_port,
  app:: 'yugabyte',
};

local yugabyteLB(metadata, name, ip) =
    if metadata.cloud_provider == "google" then googleYugabyteLB(metadata, name, ip)
    else if metadata.cloud_provider == "aws" then awsYugabyteLB(metadata, name, ip)
    else if metadata.cloud_provider == "minikube" then minikubeYugabyteLB(metadata, name, ip);
{
  all(metadata): if metadata.datastore == 'yugabyte' then {
  } else {}
}
