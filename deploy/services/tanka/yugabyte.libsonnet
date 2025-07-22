local base = import 'base.libsonnet';
local util = import 'util.libsonnet';
local volumes = import 'volumes.libsonnet';

{
  all(metadata): if metadata.datastore == 'yugabyte' then {

    # Thoses pre commands are used in yugabyte deployments to make the local ip pointing to the public hostname we want to use, until https://github.com/yugabyte/yugabyte-db/issues/27367 is fixed
    local precommand_prefix = "sed -E \"/\\.svc\\.cluster\\.local/ s/^([0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+)([[:space:]]+)/\\1 $(echo \"",
    local precommand_suffix = "\" | sed 's/[\\/&]/\\\\&/g')\\2/\" /etc/hosts > /tmp/newhosts && /bin/cp /tmp/newhosts /etc/hosts && \\",
    local master_precommand = if metadata.yugabyte.fix_27367_issue then precommand_prefix + metadata.yugabyte.master.server_broadcast_addresses + precommand_suffix else "",
    local tserver_precommand = if metadata.yugabyte.fix_27367_issue then precommand_prefix + metadata.yugabyte.tserver.server_broadcast_addresses + precommand_suffix else "",
    Master: base.StatefulSet(metadata, 'yb-master') {
      metadata+: {
        namespace: metadata.namespace,
      },
      spec+: {
        serviceName: 'yb-masters',
        podManagementPolicy: 'Parallel',
        volumeClaimTemplates: [{
          metadata: {
            name: 'datadir0',
          },
          spec: {
            storageClassName: metadata.yugabyte.storageClass,
            accessModes: [
              'ReadWriteOnce',
            ],
            resources: {
              requests: {
                storage: '10Gi',
              },
            },
          },
        }, {
          metadata: {
              name: 'datadir1',
          },
          spec: {
            storageClassName: metadata.yugabyte.storageClass,
            accessModes: [
              'ReadWriteOnce',
            ],
            resources: {
              requests: {
                storage: '10Gi',
              },
            },
          },
        }],
        template+: {
          metadata+: {
            labels+: {
              yugabytedUi: "true",
            },
          },
          spec+: {
            affinity: {
              podAntiAffinity: {},
            },
            volumes: [{
              name: "debug-hooks-volume",
              configMap: {
                name: "dss-dss-yugabyte-master-hooks",
                defaultMode: 493, # 0755
              },
            }, {
              name: "master-gflags",
              secret: {
                secretName: "dss-dss-yugabyte-master-gflags",
                defaultMode: 493, # 0755
              },
            }, {
              name: "yb-master-yugabyte-tls-cert",
              secret:{
                secretName: "yb-master-yugabyte-tls-cert",
                defaultMode: 256, # 400
              },
            }, {
              name: "yugabyte-tls-client-cert",
              secret: {
                secretName: "yugabyte-tls-client-cert",
                defaultMode: 256, # 400
              },
            }],
            containers: [
              base.Container('yb-master') {
                image: metadata.yugabyte.image,
                resources: {
                  limits: {
                    cpu: if metadata.yugabyte.light_resources then 0.1 else 2,
                    memory: if metadata.yugabyte.light_resources then "0.5Gi" else "2Gi",
                  },
                  requests: {
                    cpu: if metadata.yugabyte.light_resources then 0.1 else 2,
                    memory: if metadata.yugabyte.light_resources then "0.5Gi" else "2Gi",
                  },
                },
                ports: [{
                  name: 'http-ui',
                  containerPort: 7000,
                }, {
                  name: 'tcp-rpc-port',
                  containerPort: 7100,
                }, {
                  name: 'yugabyted-ui',
                  containerPort: 15433,
                }],
                env: [{
                  name: 'POD_IP',
                  valueFrom: {
                    fieldRef: {
                      fieldPath: "status.podIP",
                    },
                  },
                }, {
                  name: 'HOSTNAME',
                  valueFrom: {
                    fieldRef: {
                      fieldPath: "metadata.name",
                    },
                  },
                }, {
                  name: 'HOSTNAMENO',
                  valueFrom: {
                    fieldRef: {
                      fieldPath: "metadata.labels['apps.kubernetes.io/pod-index']",
                    },
                  },
                }, {
                  name: 'NAMESPACE',
                  valueFrom: {
                    fieldRef: {
                      fieldPath: "metadata.namespace",
                    },
                  },
                }, {
                  name: 'YBDEVOPS_CORECOPY_DIR',
                  value: "/mnt/disk0/cores",
                }],
                livenessProbe: {
                  exec: {
                    command: [
                      'bash',
                      '-v',
                      '-c',
                      |||
                      echo "disk check at: $(date)" \
                        | tee "/mnt/disk0/disk.check" "/mnt/disk1/disk.check" \
                        && sync "/mnt/disk0/disk.check" "/mnt/disk1/disk.check";
                      exit_code="$?";
                      echo "disk check exited with: ${exit_code}";
                      exit "${exit_code}"
                    |||
                    ],
                  },
                  failureThreshold: 3,
                  periodSeconds: 10,
                  successThreshold: 1,
                  timeoutSeconds: 1,
                },
                lifecycle: {
                  postStart: {
                    exec: {
                    command: [
                      'bash',
                      '-c',
                      |||
                        mkdir -p /mnt/disk0/cores;
                        mkdir -p /mnt/disk0/yb-data/scripts;
                        if [ ! -f /mnt/disk0/yb-data/scripts/log_cleanup.sh ]; then
                          if [ -f /home/yugabyte/bin/log_cleanup.sh ]; then
                            cp /home/yugabyte/bin/log_cleanup.sh /mnt/disk0/yb-data/scripts;
                          fi;
                        fi
                    |||
                    ],
                    },
                  },
                },
                workingDir: "/mnt/disk0/cores",
                command: [
                  "/sbin/tini",
                  "--",
                ],
                args: [
                  "/bin/bash",
                  "-c",
                  |||
                  %s
                  echo "disk check at: $(date)" \
                    | tee "/mnt/disk0/disk.check" "/mnt/disk1/disk.check" \
                    && sync "/mnt/disk0/disk.check" "/mnt/disk1/disk.check" && \
                  if [ -f /home/yugabyte/tools/k8s_preflight.py ]; then
                    PYTHONUNBUFFERED="true" /home/yugabyte/tools/k8s_preflight.py \
                      dnscheck \
                      --addr="${HOSTNAME}.yb-masters.${NAMESPACE}.svc.cluster.local" \
                      --port="7100"
                  fi && \

                  if [ -f /home/yugabyte/tools/k8s_preflight.py ]; then
                    PYTHONUNBUFFERED="true" /home/yugabyte/tools/k8s_preflight.py \
                      dnscheck \
                      --addr="${HOSTNAME}.yb-masters.${NAMESPACE}.svc.cluster.local:7100" \
                      --port="7100"
                  fi && \

                  if [ -f /home/yugabyte/tools/k8s_preflight.py ]; then
                    PYTHONUNBUFFERED="true" /home/yugabyte/tools/k8s_preflight.py \
                      dnscheck \
                      --addr="0.0.0.0" \
                      --port="7000"
                  fi && \

                  if [[ -f /home/yugabyte/tools/k8s_parent.py ]]; then
                    k8s_parent="/home/yugabyte/tools/k8s_parent.py"
                  else
                    k8s_parent=""
                  fi && \
                  mkdir -p /tmp/yugabyte/master/conf && \
                  envsubst < /opt/master/conf/server.conf.template > /tmp/yugabyte/master/conf/server.conf && \
                  exec ${k8s_parent} /home/yugabyte/bin/yb-master \
                    --flagfile /tmp/yugabyte/master/conf/server.conf
                ||| % [ master_precommand ],
                ],
                volumeMounts: [{
                  name: "master-gflags",
                  mountPath: "/opt/master/conf",
                }, {
                  name: "debug-hooks-volume",
                  mountPath: "/opt/debug_hooks_config",
                }, {
                  name: "datadir0",
                  mountPath: "/mnt/disk0",
                }, {
                  name: "datadir1",
                  mountPath: "/mnt/disk1",
                }, {
                  name: "yb-master-yugabyte-tls-cert",
                  mountPath: "/opt/certs/yugabyte",
                  readOnly: true,
                }, {
                  name: "yugabyte-tls-client-cert",
                  mountPath: "/root/.yugabytedb/",
                  readOnly: true,
                }],
              },
              base.Container('yb-cleanup') {
                image: metadata.yugabyte.image,
                env: [{
                  name: 'USER',
                  value: "yugabyte",
                }],
                command: [
                  "/sbin/tini",
                  "--",
                ],
                args: [
                  "/bin/bash",
                  "-c",
                  |||
                  while true; do
                    sleep 3600;
                    /home/yugabyte/scripts/log_cleanup.sh;
                  done
                |||,
                ],
                volumeMounts: [{
                  name: "datadir0",
                  mountPath: "/home/yugabyte/",
                  subPath: "yb-data",
                }, {
                  name: "datadir0",
                  mountPath: "/home/yugabyte/cores",
                  subPath: "cores",
                }],
              },
              base.Container('yugabyted-ui') {
                image: metadata.yugabyte.image,
                env: [{
                  name: 'POD_IP',
                  valueFrom: {
                    fieldRef: {
                      fieldPath: "status.podIP",
                    },
                  },
                }, {
                  name: 'HOSTNAME',
                  valueFrom: {
                    fieldRef: {
                      fieldPath: "metadata.name",
                    },
                  },
                }, {
                  name: 'NAMESPACE',
                  valueFrom: {
                    fieldRef: {
                      fieldPath: "metadata.namespace",
                    },
                  },
                }],
                command: [
                  "/sbin/tini",
                  "--",
                ],
                args: [
                  "/bin/bash",
                  "-c",
                  |||
                  while true; do
                    /home/yugabyte/bin/yugabyted-ui \
                      -database_host=${HOSTNAME}.yb-masters.${NAMESPACE}.svc.cluster.local \
                      -bind_address=0.0.0.0 \
                      -ysql_port=5433 \
                      -ycql_port=9042 \
                      -master_ui_port=7000 \
                      -tserver_ui_port=9000 \
                      -secure=true \
                    || echo "ERROR: yugabyted-ui failed. This might be because your yugabyte \
                    version is older than 2.21.0. If this is the case, set yugabytedUi.enabled to false \
                    in helm to disable yugabyted-ui, or upgrade to a version 2.21.0 or newer."; \
                    echo "Attempting restart in 30s."
                    trap break TERM INT; \
                    sleep 30s & wait; \
                    trap - TERM INT;
                  done \
                |||,
                ],
              },
            ],
            terminationGracePeriodSeconds: 300,
          },
        },
      },
    },

    Tserver: base.StatefulSet(metadata, 'yb-tserver') {
      metadata+: {
        namespace: metadata.namespace,
      },
      spec+: {
        serviceName: 'yb-tservers',
        podManagementPolicy: 'Parallel',
        volumeClaimTemplates: [{
          metadata: {
            name: 'datadir0',
          },
          spec: {
            storageClassName: metadata.yugabyte.storageClass,
            accessModes: [
              'ReadWriteOnce',
            ],
            resources: {
              requests: {
                storage: '10Gi',
              },
            },
          },
        }, {
          metadata: {
              name: 'datadir1',
          },
          spec: {
            storageClassName: metadata.yugabyte.storageClass,
            accessModes: [
              'ReadWriteOnce',
            ],
            resources: {
              requests: {
                storage: '10Gi',
              },
            },
          },
        }],
        template+: {
          metadata+: {
            labels+: {
              yugabytedUi: "true",
            },
          },
          spec+: {
            affinity: {
              podAntiAffinity: {},
            },
            volumes: [{
              name: "debug-hooks-volume",
              configMap: {
                name: "dss-dss-yugabyte-tserver-hooks",
                defaultMode: 493, # 0755
              },
            }, {
              name: "tserver-gflags",
              secret: {
                secretName: "dss-dss-yugabyte-tserver-gflags",
                defaultMode: 493, # 0755
              },
            }, {
              name: "yb-tserver-yugabyte-tls-cert",
              secret:{
                secretName: "yb-tserver-yugabyte-tls-cert",
                defaultMode: 256, # 400
              },
            }, {
              name: "yugabyte-tls-client-cert",
              secret: {
                secretName: "yugabyte-tls-client-cert",
                defaultMode: 256, # 400
              },
            }, {
              name: "tserver-tmp",
              emptyDir: {},
            }],
            containers: [
              base.Container('yb-tserver') {
                image: metadata.yugabyte.image,
                resources: {
                  limits: {
                    cpu: if metadata.yugabyte.light_resources then 0.1 else 2,
                    memory: if metadata.yugabyte.light_resources then "0.5Gi" else "4Gi",
                  },
                  requests: {
                    cpu: if metadata.yugabyte.light_resources then 0.1 else 2,
                    memory: if metadata.yugabyte.light_resources then "0.5Gi" else "4Gi",
                  },
                },
                ports: [{
                  name: 'http-ui',
                  containerPort: 9000,
                }, {
                  name: 'tcp-rpc-port',
                  containerPort: 9100,
                }, {
                  name: 'yugabyted-ui',
                  containerPort: 15433,
                }, {
                  name: 'http-ycql-met',
                  containerPort: 12000,
                }, {
                  name: 'http-yedis-met',
                  containerPort: 11000,
                }, {
                  name: 'http-ysql-met',
                  containerPort: 13000,
                }, {
                  name: 'tcp-yedis-port',
                  containerPort: 6379,
                }, {
                  name: 'tcp-yql-port',
                  containerPort: 9042,
                }, {
                  name: 'tcp-ysql-port',
                  containerPort: 5433,
                }],
                env: [{
                  name: 'POD_IP',
                  valueFrom: {
                    fieldRef: {
                      fieldPath: "status.podIP",
                    },
                  },
                }, {
                  name: 'HOSTNAME',
                  valueFrom: {
                    fieldRef: {
                      fieldPath: "metadata.name",
                    },
                  },
                }, {
                  name: 'HOSTNAMENO',
                  valueFrom: {
                    fieldRef: {
                      fieldPath: "metadata.labels['apps.kubernetes.io/pod-index']",
                    },
                  },
                }, {
                  name: 'NAMESPACE',
                  valueFrom: {
                    fieldRef: {
                      fieldPath: "metadata.namespace",
                    },
                  },
                }, {
                  name: 'YBDEVOPS_CORECOPY_DIR',
                  value: "/mnt/disk0/cores",
                }, {
                  name: 'SSL_CERTFILE',
                  value: "/root/.yugabytedb/root.crt",
                }],
                livenessProbe: {
                  exec: {
                    command: [
                      'bash',
                      '-v',
                      '-c',
                      |||
                      echo "disk check at: $(date)" \
                        | tee "/mnt/disk0/disk.check" "/mnt/disk1/disk.check" \
                        && sync "/mnt/disk0/disk.check" "/mnt/disk1/disk.check";
                      exit_code="$?";
                      echo "disk check exited with: ${exit_code}";
                      exit "${exit_code}"
                    |||
                    ],
                  },
                  failureThreshold: 3,
                  periodSeconds: 10,
                  successThreshold: 1,
                  timeoutSeconds: 1,
                },
                lifecycle: {
                  postStart: {
                    exec: {
                    command: [
                      'bash',
                      '-c',
                      |||
                        mkdir -p /mnt/disk0/cores;
                        mkdir -p /mnt/disk0/yb-data/scripts;
                        if [ ! -f /mnt/disk0/yb-data/scripts/log_cleanup.sh ]; then
                          if [ -f /home/yugabyte/bin/log_cleanup.sh ]; then
                            cp /home/yugabyte/bin/log_cleanup.sh /mnt/disk0/yb-data/scripts;
                          fi;
                        fi
                    |||
                    ],
                    },
                  },
                },
                workingDir: "/mnt/disk0/cores",
                command: [
                  "/sbin/tini",
                  "--",
                ],
                args: [
                  "/bin/bash",
                  "-c",
                  |||
                  %s
                  echo "disk check at: $(date)" \
                    | tee "/mnt/disk0/disk.check" "/mnt/disk1/disk.check" \
                    && sync "/mnt/disk0/disk.check" "/mnt/disk1/disk.check" && \
                  if [ -f /home/yugabyte/tools/k8s_preflight.py ]; then
                    PYTHONUNBUFFERED="true" /home/yugabyte/tools/k8s_preflight.py \
                      dnscheck \
                      --addr="${HOSTNAME}.yb-tservers.${NAMESPACE}.svc.cluster.local" \
                      --port="9100"
                  fi && \

                  if [ -f /home/yugabyte/tools/k8s_preflight.py ]; then
                    PYTHONUNBUFFERED="true" /home/yugabyte/tools/k8s_preflight.py \
                      dnscheck \
                      --addr="${HOSTNAME}.yb-tservers.${NAMESPACE}.svc.cluster.local:9100" \
                      --port="9100"
                  fi && \

                  if [ -f /home/yugabyte/tools/k8s_preflight.py ]; then
                    PYTHONUNBUFFERED="true" /home/yugabyte/tools/k8s_preflight.py \
                      dnscheck \
                      --addr="0.0.0.0" \
                      --port="9000"
                  fi && \

                  if [[ -f /home/yugabyte/tools/k8s_parent.py ]]; then
                    k8s_parent="/home/yugabyte/tools/k8s_parent.py"
                  else
                    k8s_parent=""
                  fi && \
                  if [ -f /home/yugabyte/tools/k8s_preflight.py ]; then
                    PYTHONUNBUFFERED="true" /home/yugabyte/tools/k8s_preflight.py \
                      dnscheck \
                      --addr="0.0.0.0:9042" \
                      --port="9042"
                  fi && \

                  if [ -f /home/yugabyte/tools/k8s_preflight.py ]; then
                    PYTHONUNBUFFERED="true" /home/yugabyte/tools/k8s_preflight.py \
                      dnscheck \
                      --addr="0.0.0.0:5433" \
                      --port="5433"
                  fi && \

                    mkdir -p /tmp/yugabyte/tserver/conf && \
                    envsubst < /opt/tserver/conf/server.conf.template > /tmp/yugabyte/tserver/conf/server.conf && \
                    exec ${k8s_parent} /home/yugabyte/bin/yb-tserver \
                      --flagfile /tmp/yugabyte/tserver/conf/server.conf
                ||| % [ tserver_precommand ],
                ],
                volumeMounts: [{
                  name: "tserver-tmp",
                  mountPath: "/tmp",
                }, {
                  name: "tserver-gflags",
                  mountPath: "/opt/tserver/conf",
                }, {
                  name: "debug-hooks-volume",
                  mountPath: "/opt/debug_hooks_config",
                }, {
                  name: "datadir0",
                  mountPath: "/mnt/disk0",
                }, {
                  name: "datadir1",
                  mountPath: "/mnt/disk1",
                }, {
                  name: "yb-tserver-yugabyte-tls-cert",
                  mountPath: "/opt/certs/yugabyte",
                  readOnly: true,
                }, {
                  name: "yugabyte-tls-client-cert",
                  mountPath: "/root/.yugabytedb/",
                  readOnly: true,
                }],
              },
              base.Container('yb-cleanup') {
                image: metadata.yugabyte.image,
                env: [{
                  name: 'USER',
                  value: "yugabyte",
                }],
                command: [
                  "/sbin/tini",
                  "--",
                ],
                args: [
                  "/bin/bash",
                  "-c",
                  |||
                  while true; do
                    sleep 3600;
                    /home/yugabyte/scripts/log_cleanup.sh;
                  done
                |||,
                ],
                volumeMounts: [{
                  name: "datadir0",
                  mountPath: "/home/yugabyte/",
                  subPath: "yb-data",
                }, {
                  name: "datadir0",
                  mountPath: "/var/yugabyte/cores",
                  subPath: "cores",
                }],
              },
              base.Container('yugabyted-ui') {
                image: metadata.yugabyte.image,
                env: [{
                  name: 'POD_IP',
                  valueFrom: {
                    fieldRef: {
                      fieldPath: "status.podIP",
                    },
                  },
                }, {
                  name: 'HOSTNAME',
                  valueFrom: {
                    fieldRef: {
                      fieldPath: "metadata.name",
                    },
                  },
                }, {
                  name: 'NAMESPACE',
                  valueFrom: {
                    fieldRef: {
                      fieldPath: "metadata.namespace",
                    },
                  },
                }],
                command: [
                  "/sbin/tini",
                  "--",
                ],
                args: [
                  "/bin/bash",
                  "-c",
                  |||
                  while true; do
                    /home/yugabyte/bin/yugabyted-ui \
                      -database_host=${HOSTNAME}.yb-tservers.${NAMESPACE}.svc.cluster.local \
                      -bind_address=0.0.0.0 \
                      -ysql_port=5433 \
                      -ycql_port=9042 \
                      -master_ui_port=7000 \
                      -tserver_ui_port=9000 \
                      -secure=true \
                    || echo "ERROR: yugabyted-ui failed. This might be because your yugabyte \
                    version is older than 2.21.0. If this is the case, set yugabytedUi.enabled to false \
                    in helm to disable yugabyted-ui, or upgrade to a version 2.21.0 or newer."; \
                    echo "Attempting restart in 30s."
                    trap break TERM INT; \
                    sleep 30s & wait; \
                    trap - TERM INT;
                  done \
                |||,
                ],
              },
            ],
            terminationGracePeriodSeconds: 300,
          },
        },
      },
    },
  } else {}
}
