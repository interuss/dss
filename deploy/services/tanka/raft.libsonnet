local base = import 'base.libsonnet';
local util = import 'util.libsonnet';
local volumes = import 'volumes.libsonnet';
local args = import 'core-service-args.libsonnet';

{
  all(metadata): {
    local raft = metadata.raft,
    local nodes = raft.nodes,
    local allNodes = nodes + raft.peers,
    local ids = [n.id for n in allNodes],
    local badEndpoints = [n.endpoint for n in allNodes
      if std.length(std.split(n.endpoint, '://')) > 1 || std.length(std.split(n.endpoint, ':')) != 2],
    local duplicateIds = [i for i in std.set(ids) if std.count(ids, i) > 1],

    // Flat values indexed by pod ordinal
    local idsList = std.join(' ', [std.toString(n.id) for n in nodes]),
    // The --raft_peers list is nodes + peers (self included), all addressed over https.
    local peersList = std.join(',', [std.toString(n.id) + '=https://' + n.endpoint for n in allNodes]),
    local ports = std.set([std.parseInt(std.split(n.endpoint, ':')[1]) for n in nodes]),

    local allArgs = args.all(metadata) + {
      store_type: 'raft',
      raft_node_id: '$(printf \'%s\\n\' $ids | sed -n "${ordinal}p")',
      raft_cluster_id: raft.clusterID,
      raft_datadir: '/raft_data',
      raft_peers: peersList,
      raft_tls_cert_file: '/raft-tls/tls.crt',
      raft_tls_key_file: '/raft-tls/tls.key',
      raft_tls_ca_file: '/raft-tls/ca-bundle.crt',
    },
    local script = std.join('\n', [
      '# Raft node IDs of this instance\'s nodes, in pod-ordinal order.',
      'ids="' + idsList + '"',
      'ordinal=$(( ${HOSTNAME##*-} + 1 ))',
      'exec core-service ' + std.join(' \\\n  ',
        std.map(function(kv)
          if kv[0] == 'raft_node_id' then '--raft_node_id=' + kv[1]
          else '\'' + ('--%s=%s' % kv) + '\'',
          util.objectItems(allArgs))),
    ]),

    assert std.length(nodes) > 0 : 'raft.nodes must list at least one node',
    assert badEndpoints == [] : 'raft node endpoints must be host:port without a scheme: ' + std.join(', ', badEndpoints),
    assert duplicateIds == [] : 'duplicate Raft node IDs in raft.nodes / raft.peers: ' + std.join(', ', [std.toString(i) for i in duplicateIds]),
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
