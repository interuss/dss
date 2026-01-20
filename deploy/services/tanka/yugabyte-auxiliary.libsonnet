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
      // Replicated from the official Helm chart with all command-line flags
      MasterGFlags: base.Secret(metadata, 'dss-dss-yugabyte-master-gflags') {
        type: "Opaque",
        stringData: {
          "server.conf.template": |||
            --fs_data_dirs=/mnt/disk0,/mnt/disk1
            --master_addresses=%s
            --replication_factor=%s
            --enable_ysql=true
            --certs_dir=/opt/certs/yugabyte
            --use_node_to_node_encryption=true
            --allow_insecure_connections=false
            --master_enable_metrics_snapshotter=true
            --metrics_snapshotter_tserver_metrics_whitelist=handler_latency_yb_tserver_TabletServerService_Read_count,handler_latency_yb_tserver_TabletServerService_Write_count,handler_latency_yb_tserver_TabletServerService_Read_sum,handler_latency_yb_tserver_TabletServerService_Write_sum,disk_usage,cpu_usage,node_up
            --metric_node_name=${EXPORTED_INSTANCE}
            --memory_limit_hard_bytes=1824522240
            --stderrthreshold=0
            --num_cpus=2
            --max_log_size=256
            --undefok=num_cpus,enable_ysql
            --rpc_bind_addresses=%s
            --server_broadcast_addresses=%s
            --webserver_interface=0.0.0.0
            --default_memory_limit_to_ram_ratio=0.85
            --placement_cloud=%s
            --placement_region=%s
            --placement_zone=%s
            --use_private_ip=zone
            --node_to_node_encryption_use_client_certificates=true
          ||| % [
            std.join(",", metadata.yugabyte.masterAddresses),
            std.length(metadata.yugabyte.masterAddresses),
            metadata.yugabyte.master.rpc_bind_addresses,
            metadata.yugabyte.master.server_broadcast_addresses,
            metadata.yugabyte.placement.cloud,
            metadata.yugabyte.placement.region,
            metadata.yugabyte.placement.zone,
          ]
        }
      },
      TServerGFlags: base.Secret(metadata, 'dss-dss-yugabyte-tserver-gflags') {
        type: "Opaque",
        stringData: {
          "server.conf.template": |||
            --fs_data_dirs=/mnt/disk0,/mnt/disk1
            --tserver_master_addrs=%s
            --certs_dir=/opt/certs/yugabyte
            --use_node_to_node_encryption=true
            --allow_insecure_connections=false
            --use_client_to_server_encryption=true
            --certs_for_client_dir=/opt/certs/yugabyte
            --cert_node_filename=${HOSTNAME}.yb-tservers.${NAMESPACE}.svc.cluster.local
            --enable_ysql=true
            --pgsql_proxy_bind_address=0.0.0.0:5433
            --tserver_enable_metrics_snapshotter=true
            --metrics_snapshotter_interval_ms=11000
            --metrics_snapshotter_tserver_metrics_whitelist=handler_latency_yb_tserver_TabletServerService_Read_count,handler_latency_yb_tserver_TabletServerService_Write_count,handler_latency_yb_tserver_TabletServerService_Read_sum,handler_latency_yb_tserver_TabletServerService_Write_sum,disk_usage,cpu_usage,node_up
            --metric_node_name=${EXPORTED_INSTANCE}
            --memory_limit_hard_bytes=3649044480
            --stderrthreshold=0
            --max_log_size=256
            --num_cpus=2
            --undefok=num_cpus,enable_ysql
            --use_node_hostname_for_local_tserver=true
            --cql_proxy_bind_address=0.0.0.0:9042
            --rpc_bind_addresses=%s
            --server_broadcast_addresses=%s
            --webserver_interface=0.0.0.0
            --placement_cloud=%s
            --placement_region=%s
            --placement_zone=%s
            --use_private_ip=zone
            --node_to_node_encryption_use_client_certificates=true
            --ysql_hba_conf_csv=hostssl all all 0.0.0.0/0 cert
          ||| % [
            std.join(",", metadata.yugabyte.masterAddresses),
            metadata.yugabyte.tserver.rpc_bind_addresses,
            metadata.yugabyte.tserver.server_broadcast_addresses,
            metadata.yugabyte.placement.cloud,
            metadata.yugabyte.placement.region,
            metadata.yugabyte.placement.zone,
          ]
        }
      },
      MasterHooks: base.ConfigMap(metadata, 'dss-dss-yugabyte-master-hooks') {
        data: {
          ['yb-master-%s-pre_debug_hook.sh' % id]: "echo 'hello-from-pre' "
          for id in std.range(0, std.length(metadata.yugabyte.masterNodeIPs) - 1)
        } + {
          ['yb-master-%s-post_debug_hook.sh' % id]: "echo 'hello-from-post' "
          for id in std.range(0, std.length(metadata.yugabyte.masterNodeIPs) - 1)
        }
      },
      TServerHooks: base.ConfigMap(metadata, 'dss-dss-yugabyte-tserver-hooks') {
        data: {
          ['yb-tserver-%s-pre_debug_hook.sh' % id]: "echo 'hello-from-pre' "
          for id in std.range(0, std.length(metadata.yugabyte.tserverNodeIPs) - 1)
        } + {
          ['yb-tserver-%s-post_debug_hook.sh' % id]: "echo 'hello-from-post' "
          for id in std.range(0, std.length(metadata.yugabyte.tserverNodeIPs) - 1)
        }
      },
      masters: base.Service(metadata, 'yb-masters') {
        app:: 'yb-master',
        spec+: {
          clusterIP: "None",
          ports: [
            {
              port: 7000,
              name: 'http-ui',
            },
            {
              port: 7100,
              name: 'tcp-rpc-port',
            },
            {
              port: 15433,
              name: 'yugabyted-ui',
            },
          ],
        },
      },
      tServers: base.Service(metadata, 'yb-tservers') {
        app:: 'yb-tserver',
        spec+: {
          clusterIP: "None",
          ports: [
            {
              port: 9000,
              name: 'http-ui',
            },
            {
              port: 12000,
              name: 'http-ycql-met',
            },
            {
              port: 11000,
              name: 'http-yedis-met',
            },
            {
              port: 13000,
              name: 'http-ysql-met',
            },
            {
              port: 9100,
              name: 'tcp-rpc-port',
            },
            {
              port: 6379,
              name: 'tcp-yedis-port',
            },
            {
              port: 9042,
              name: 'tcp-yql-port',
            },
            {
              port: 5433,
              name: 'tcp-ysql-port',
            },
            {
              port: 15433,
              name: 'yugabyted-ui',
            },
          ],
        },
      },
      individualMasters: {
        ["master-" + i]: base.Service(metadata, 'yb-master-' + i) {
          app:: 'yb-master-' + i,
          metadata+: {
            annotations+: {
              'prometheus.io/scrape': 'true',
              'prometheus.io/port': '7000',
              'prometheus.io/path': 'prometheus-metrics',
            },
          },
          spec+: {
            selector: {
              'app': "yb-master",
              'apps.kubernetes.io/pod-index': std.toString(i),
            },
            clusterIP: "None",
            ports: [
              {
                port: 7000,
                name: 'http-ui',
              },
            ],
          },
        } for i in std.range(0, std.length(metadata.yugabyte.masterNodeIPs) - 1)
      },
      individualTServers: {
        ["tserver-" + i]: base.Service(metadata, 'yb-tserver-' + i) {
          app:: 'yb-tserver-' + i,
          metadata+: {
            annotations+: {
              'prometheus.io/scrape': 'true',
              'prometheus.io/port': '9000',
              'prometheus.io/path': 'prometheus-metrics',
            },
          },
          spec+: {
            selector: {
              'app': "yb-tserver",
              'apps.kubernetes.io/pod-index': std.toString(i),
            },
            clusterIP: "None",
            ports: [
              {
                port: 9000,
                name: 'http-ui',
              },
            ],
          },
        } for i in std.range(0, std.length(metadata.yugabyte.masterNodeIPs) - 1)
      },
      individualTServersYsql: {
        ["ysql-" + i]: base.Service(metadata, 'yb-ysql-' + i) {
          app:: 'yb-ysql-' + i,
          metadata+: {
            annotations+: {
              'prometheus.io/scrape': 'true',
              'prometheus.io/port': '13000',
              'prometheus.io/path': 'prometheus-metrics',
            },
          },
          spec+: {
            selector: {
              'app': "yb-tserver",
              'apps.kubernetes.io/pod-index': std.toString(i),
            },
            clusterIP: "None",
            ports: [
              {
                port: 13000,
                name: 'http-ysql-met',
              },
            ],
          },
        } for i in std.range(0, std.length(metadata.yugabyte.masterNodeIPs) - 1)
      },
      NodeGateways: {
        ["gateway-" + i]: yugabyteLB(metadata, 'ybdb-ext-' + i, metadata.yugabyte.masterNodeIPs[i]) {
          metadata+: {
            annotations+: {
              'service.alpha.kubernetes.io/tolerate-unready-endpoints': 'true',
            },
          },
          spec+: {
            selector: {
              app: 'yugabyte-proxy-' + i
            },
            publishNotReadyAddresses: true,
            ports: [
              {
                port: 7100,
                name: 'tcp-rpc-port',
              },
              {
                port: 9100,
                name: 'tcp-rpc2-port',
              },
            ],
          },
        } for i in std.range(0, std.length(metadata.yugabyte.tserverNodeIPs) - 1)
      },
      ProxyConfig: {
        ["yb-proxy-config-" + i]: base.ConfigMap(metadata, 'yb-proxy-config-' + i) {
          data: {
            "haproxy.cfg": |||
              global
                log stdout format raw local0
                maxconn 4096

              defaults
                mode tcp
                log global
                # We set high timeouts to avoid disconnects with low activitiy
                timeout client 12h
                timeout server 12h
                timeout tunnel 12h
                timeout connect 5s
                # We enable TCP keep alives on client and server side
                option clitcpka
                option srvtcpka
                # K8s services may not be ready when HaProxy start, we ignore errors
                default-server init-addr libc,none

              resolvers dns
                parse-resolv-conf
                # We limit DNS validity to 5s to react to changes on K8s services
                hold valid 5s

              frontend master-grpc-f
                bind :7100
                default_backend master-grpc-b

              backend master-grpc-b
                server yb-master-%s yb-master-%s.yb-masters.%s.svc.cluster.local:7100 check resolvers dns

              frontend tserver-grpc-f
                bind :9100
                default_backend tserver-grpc-b

              backend tserver-grpc-b
                server yb-tserver-%s yb-tserver-%s.yb-tservers.%s.svc.cluster.local:9100 check resolvers dns
          ||| % [i, i, metadata.namespace, i, i, metadata.namespace]
          }
        } for i in std.range(0, std.length(metadata.yugabyte.tserverNodeIPs) - 1)
      },
      Proxy: {
        ["yugabyte-proxy-" + i]: base.Deployment(metadata, 'yugabyte-proxy-' + i) {
          apiVersion: 'apps/v1',
          kind: 'Deployment',
          metadata+: {
            namespace: metadata.namespace,
            labels: {
              name: 'yugabyte-proxy-' + i
            }
          },
          spec+: {
            replicas: 2,  # We deploy two instances to provide resilience if one nodes goes down.
            selector: {
              matchLabels: {
                app: 'yugabyte-proxy-' + i
              }
            },
            strategy: {
              rollingUpdate: {
                maxSurge: "25%",
                maxUnavailable: "25%",
              },
              type: "RollingUpdate",
            },
            template+: {
              metadata+: {
                labels: {
                  app: 'yugabyte-proxy-' + i
                }
              },
              spec+: {
                volumes: [
                  {
                    name: "config-volume",
                    configMap: {
                      name: "yb-proxy-config-" + i,
                    }
                  }
                ],
                soloContainer:: base.Container('yugabyte-proxy') {
                  image: "haproxy:3.3",
                  imagePullPolicy: 'Always',
                  ports: [
                    {
                      containerPort: 7100,
                      name: 'master-grpc',
                    },
                    {
                      containerPort: 9100,
                      name: 'tserver-grpc',
                    },
                  ],
                  volumeMounts: [
                    {
                      name: "config-volume",
                      mountPath: "/usr/local/etc/haproxy/",
                    }
                  ],
                },
              },
            },
          },
        } for i in std.range(0, std.length(metadata.yugabyte.tserverNodeIPs) - 1)
    },
  } else {}
}
