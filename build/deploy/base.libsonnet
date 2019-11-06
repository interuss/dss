local util = import 'util.libsonnet';

{
  _Object(apiVersion, kind, name, metadata):: {
    apiVersion: apiVersion,
    kind: kind,
    metadata: {
      name: name,
      namespace: metadata.namespace,
      labels: { name: std.join('-', std.split(name, ':')) },
    },
  },

  Deployment(metadata, name): $._Object('apps/v1', 'Deployment', name, metadata) {
    local deployment = self,
    app:: error "must specify app",

    spec: {
      template: {
        metadata+: {
          labels+: {
            app: deployment.app,
          },
        },
        spec: $.PodSpec,
      },

      strategy: {
        type: 'RollingUpdate',

        rollingUpdate: {
          maxSurge: '25%',  // rounds up
          maxUnavailable: '25%',  // rounds down
        },
      },

      minReadySeconds: 30,

      replicas: 1,
    },
  },

  Service(metadata, name): $._Object('v1', 'Service', name, metadata) {
    local service = self,
    app:: error "must specify app",
    # Helper when a service has only 1 port
    port:: error "must specify port",
    enable_monitoring:: false,
    type:: 'ClusterIP',
    metadata+: {
      labels+: {
        app: service.app,
      },
      annotations: if service.enable_monitoring then {
        'prometheus.io/scrape': 'true',
        'prometheus.io/path': '_status/vars',
        'prometheus.io/port': service.port,
      } else {},
    },
    spec: {
      selector: {
        app: service.app,
      },
      ports: [
        {
          port: service.port,
          targetPort: service.port,
          name: name,
        },
      ],
      type: service.type,
    },
  },

  StatefulSet(metadata, name): $._Object('apps/v1', 'StatefulSet', name, metadata) {
    local sset = self,

    spec: {
      serviceName: name,

      updateStrategy: {
        type: 'RollingUpdate',
        rollingUpdate: {
          partition: 0,
        },
      },

      template: {
        spec: $.PodSpec,
        metadata: {
          labels: sset.metadata.labels,
          annotations: {},
        },
      },

      selector: {
        matchLabels: sset.spec.template.metadata.labels,
      },

      volumeClaimTemplates_:: {},
      volumeClaimTemplates: [
        // StatefulSet is overly fussy about "changes" (even when
        // they're no-ops).
        // In particular annotations={} is apparently a "change",
        // since the comparison is ignorant of defaults.
        std.prune($.PersistentVolumeClaim($.hyphenate(kv[0])) + { apiVersion:: null, kind:: null } + kv[1])
        for kv in util.objectItems(self.volumeClaimTemplates_)
      ],

      replicas: 1,
      assert self.replicas >= 1,
    },
  },

  Job(metadata, name): $._Object('batch/v1', 'Job', name, metadata) {
    local job = self,

    spec: {
      template: {
        metadata+: {
          labels: job.metadata.labels,
        },
        spec: $.PodSpec {
          restartPolicy: 'OnFailure',
        },
      },
      completions: 1,
      parallelism: 1,
    },
  },

  Container(name): {
    name: name,
    image: error 'container image value required',
    imagePullPolicy: if std.endsWith(self.image, ':latest') then 'Always' else 'IfNotPresent',

    args_:: {},
    args: ['--%s=%s' % kv for kv in util.objectItems(self.args_)],

    stdin: false,
    tty: false,
    assert !self.tty || self.stdin : 'tty=true requires stdin=true',
  },

  PodSpec: {
    soloContainer:: error 'must have at least one container',
    containers: [self.soloContainer],
    volumes_:: {},
    volumes: util.mapToList(self.volumes_),

    terminationGracePeriodSeconds: 30,

    assert std.length(self.containers) > 0 : 'must have at least one container',
  },
}