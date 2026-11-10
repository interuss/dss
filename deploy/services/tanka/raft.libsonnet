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
            volumes: volumes.all(metadata).backendVolumes,
            soloContainer:: base.Container('core-service') {
              image: metadata.backend.image,
              imagePullPolicy: if metadata.backend.imagePullPolicy != '' then metadata.backend.imagePullPolicy
                else if metadata.cloud_provider == 'minikube' then 'IfNotPresent' else 'Always',
              command: ['sh', '-c'],
              args: [script],
              ports: [
                { containerPort: metadata.backend.port, name: 'http' },
              ] + (if metadata.enableDssMetrics then [{ containerPort: 8079, name: 'metrics' }] else [])
              + [{ containerPort: p } for p in ports],
              volumeMounts: volumes.all(metadata).backendMounts,
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
