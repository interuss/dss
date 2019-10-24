// kube.libsonnet is an import from bitnami, we would not maintain this import this way.
local kube = import "kube.libsonnet"; 
local common = import "common.libsonnet";

local cockroachLB(name, ip) = kube.Service(name) {
  spec: {
    ports: [
      {
        port: common.cockroach.grpc_port,
        targetPort: common.cockroach.grpc_port,
      },
    ],
    selector: error "must specify selector",
    type: "LoadBalancer",
    loadBalancerIP: ip,
  },
};

// Object to return
{
  Balanced(ip): cockroachLB("cockroach-db-external-balanced", ip) {
      spec+: {
        selector: {
          app: "cockroachdb",
        },
      },
  },

  NodeGateways(ip_list): [
    cockroachLB("cockroach-db-external-node-" + i, ip_list[i]) {
      spec+: {
        selector: {
          "statefulset.kubernetes.io/pod-name": "cockroachdb-" + i,
        },  
      },
    } for i in std.range(0, std.length(ip_list) - 1)
  ],
}