local base = import 'base.libsonnet';
local util = import 'util.libsonnet';
local volumes = import 'volumes.libsonnet';
local args = import 'core-service-args.libsonnet';

{
  all(metadata): {
    local raft = metadata.raft,
    local nodes = raft.nodes,
    local endpoints = [n.endpoint for n in nodes],
    local badEndpoints = [e for e in endpoints
      if std.length(std.split(e, '://')) > 1 || std.length(std.split(e, ':')) != 2],
    local duplicateEndpoints = [e for e in std.set(endpoints) if std.count(endpoints, e) > 1],

    local ports = std.set([std.parseInt(std.split(e, ':')[1]) for e in endpoints]),

    local allArgs = args.all(metadata) + {
      store_type: 'raft',
      raft_peer_url: '"https://$(printf \'%s\\n\' $endpoints | sed -n "${ordinal}p")"',
      raft_cluster_id: raft.clusterID,
      raft_datadir: '/raft_data',
      raft_tls_cert_file: '"/raft-tls/node.raft-${HOSTNAME##*-}.crt"',
      raft_tls_key_file: '"/raft-tls/node.raft-${HOSTNAME##*-}.key"',
      raft_tls_ca_file: '/raft-tls/ca.crt',
    },
    local script = std.join('\n', [
      '# Raft endpoints of this instance\'s nodes, in pod-ordinal order.',
      'endpoints="' + std.join(' ', endpoints) + '"',
      'ordinal=$(( ${HOSTNAME##*-} + 1 ))',
      'exec core-service ' + std.join(' \\\n  ',
        std.map(function(kv)
          if std.member(['raft_peer_url', 'raft_tls_cert_file', 'raft_tls_key_file'], kv[0]) then '--%s=%s' % kv
          else '\'' + ('--%s=%s' % kv) + '\'',
          util.objectItems(allArgs))),
    ]),

    local imagePullPolicy = if metadata.backend.imagePullPolicy != '' then metadata.backend.imagePullPolicy
      else if metadata.cloud_provider == 'minikube' then 'IfNotPresent' else 'Always',
    local adminSocketMount = { name: 'admin-socket', mountPath: '/run/dss' },
    local bootstrapScript = |||
      trap 'exit 0' TERM
      if [ "${HOSTNAME##*-}" != "0" ]; then
        sleep infinity &
        wait
      fi
      until dssctl raft node > /dev/null 2>&1; do
        sleep 5 &
        wait
      done
      for group in aux rid scd; do
        until dssctl raft node --groups "${group}" | grep -q ': member'; do
          dssctl raft bootstrap --groups "${group}" || sleep 5
        done &
      done
      wait
      endpoints="%s"
      while :; do
        pending=""
        for endpoint in ${endpoints}; do
          for group in aux rid scd; do
            if ! dssctl raft list --groups "${group}" | grep -qF "https://${endpoint} "; then
              pending="${pending} ${endpoint}"
              break
            fi
          done
        done
        if [ -z "${pending}" ]; then
          break
        fi
        endpoints="${pending}"
        reachable=$(
          for endpoint in ${pending}; do
            (
              curl -sk --max-time 1 -o /dev/null "https://${endpoint}"
              case $? in
                6|7|28) ;;
                *) echo "${endpoint}" ;;
              esac
            ) &
          done
          wait
        )
        for endpoint in ${reachable}; do
          echo "${endpoint} reachable, adding it"
        done
        for group in aux rid scd; do
          (
            for endpoint in ${reachable}; do
              if ! dssctl raft list --groups "${group}" | grep -qF "https://${endpoint} "; then
                dssctl raft add --groups "${group}" --url "https://${endpoint}"
              fi
            done
          ) &
        done
        wait
        sleep 5 &
        wait
      done
      echo "Bootstrapping complete"
      sleep infinity &
      wait
    ||| % std.join(' ', endpoints[1:] + raft.automaticBootstrapNodes),

    assert std.length(nodes) > 0 : 'raft.nodes must list at least one node',
    assert badEndpoints == [] : 'raft node endpoints must be host:port without a scheme: ' + std.join(', ', badEndpoints),
    assert duplicateEndpoints == [] : 'duplicate endpoints in raft.nodes: ' + std.join(', ', duplicateEndpoints),
    assert raft.tlsSecret != '' : 'raft.tlsSecret is required: the Raft listeners authenticate with mutual TLS',

    statefulset: base.StatefulSet(metadata, 'core-service') {
      spec+: {
        serviceName: null,
        podManagementPolicy: 'Parallel',
        minReadySeconds: 30,
        replicas: std.length(nodes),
        volumeClaimTemplates: [{
          metadata: { name: 'raft-data' },
          spec: {
            accessModes: ['ReadWriteOnce'],
            storageClassName: if raft.storage.storageClass != '' then raft.storage.storageClass else null,
            resources: { requests: { storage: raft.storage.size } },
          },
        }],
        template+: {
          metadata+: {
            annotations: if metadata.enableDssMetrics then {
              'prometheus.io/path': 'metrics',
              'prometheus.io/port': '8079',
              'prometheus.io/scrape': 'true',
            } else {},
          },
          spec+: {
            affinity: {
              podAntiAffinity: {
                preferredDuringSchedulingIgnoredDuringExecution: [{
                  weight: 100,
                  podAffinityTerm: {
                    labelSelector: { matchLabels: { app: 'core-service' } },
                    topologyKey: 'kubernetes.io/hostname',
                  },
                }],
              },
            },
            volumes: volumes.all(metadata).backendVolumes
              + (if raft.automaticBootstrap then [{ name: 'admin-socket', emptyDir: {} }] else []),
            containers: [self.soloContainer] + (if raft.automaticBootstrap then [base.Container('raft-bootstrapper') {
              image: metadata.backend.image,
              imagePullPolicy: imagePullPolicy,
              command: ['sh', '-c'],
              args: [bootstrapScript],
              volumeMounts: [adminSocketMount],
            }] else []),
            soloContainer:: base.Container('core-service') {
              image: metadata.backend.image,
              imagePullPolicy: imagePullPolicy,
              command: ['sh', '-c'],
              args: [script],
              ports: [
                { containerPort: metadata.backend.port, name: 'http' },
              ] + (if metadata.enableDssMetrics then [{ containerPort: 8079, name: 'metrics' }] else [])
              + [{ containerPort: p } for p in ports],
              volumeMounts: volumes.all(metadata).backendMounts
                + (if raft.automaticBootstrap then [adminSocketMount] else []),
              env: util.makeEnv(metadata.backend.env),
              resources: util.resources(metadata.backend),
              readinessProbe: {
                httpGet: { path: '/healthy', port: metadata.backend.port },
              },
            },
          },
        },
      },
    },
  },
}
