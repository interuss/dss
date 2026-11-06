local base = import 'base.libsonnet';

local googleRaftLB(metadata, name, ip, port) = base.Service(metadata, name) {
  port:: port,
  app:: 'core-service',
  spec+: {
    type: 'LoadBalancer',
    loadBalancerIP: ip,
  },
};

local awsRaftLB(metadata, name, ip, port) = base.AWSLoadBalancer(metadata, name, [ip], metadata.subnet) {
  port:: port,
  app:: 'core-service',
};

local minikubeRaftLB(metadata, name, ip, port) = base.Service(metadata, name) {
  port:: port,
  app:: 'core-service',
  spec+: {
    type: 'LoadBalancer',
  },
};

local raftNodeLB(metadata, i, node) =
  local endpoint = std.split(node.endpoint, ':'),
         name = 'raft-external-node-' + std.toString(i),
         ip = node.ip,
         port = std.parseInt(endpoint[1]);
  (if metadata.cloud_provider == 'google' then googleRaftLB(metadata, name, ip, port)
   else if metadata.cloud_provider == 'aws' then awsRaftLB(metadata, name, ip, port)
   else if metadata.cloud_provider == 'minikube' then minikubeRaftLB(metadata, name, ip, port)) {
    metadata+: {
      annotations+: {
        'service.alpha.kubernetes.io/tolerate-unready-endpoints': 'true',
      },
    },
    spec+: {
      selector: { 'statefulset.kubernetes.io/pod-name': 'core-service-' + std.toString(i) },
      publishNotReadyAddresses: true,
    },
  };

{
  all(metadata):
    local nodes = metadata.raft.nodes;
    {
      NodeGateways: {
        ['raft-external-node-' + std.toString(i)]: raftNodeLB(metadata, i, nodes[i])
        for i in std.range(0, std.length(nodes) - 1)
      },
    },
}
