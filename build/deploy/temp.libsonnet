{
  apiVersion: 'apps/v1beta1',
  kind: 'StatefulSet',
  metadata: {
    name: 'cockroachdb',
    namespace: {
      '[object Object]': null,
    },
  },
  spec: {
    serviceName: 'cockroachdb',
    replicas: 3,
    template: {
      metadata: {
        labels: {
          app: 'cockroachdb',
        },
      },
      spec: {
        serviceAccountName: 'cockroachdb',
        affinity: {
          podAntiAffinity: {
            preferredDuringSchedulingIgnoredDuringExecution: [
              {
                weight: 100,
                podAffinityTerm: {
                  labelSelector: {
                    matchExpressions: [
                      {
                        key: 'app',
                        operator: 'In',
                        values: [
                          'cockroachdb',
                        ],
                      },
                    ],
                  },
                  topologyKey: 'kubernetes.io/hostname',
                },
              },
            ],
          },
        },
        containers: [
          {
            name: 'cockroachdb',
            image: {
              '[object Object]': null,
            },
            imagePullPolicy: 'IfNotPresent',
            ports: [
              {
                containerPort: {
                  '[object Object]': null,
                },
                name: 'grpc',
              },
              {
                containerPort: {
                  '[object Object]': null,
                },
                name: 'http',
              },
            ],
            livenessProbe: {
              httpGet: {
                path: '/health',
                port: 'http',
                scheme: 'HTTPS',
              },
              initialDelaySeconds: 30,
              periodSeconds: 5,
            },
            readinessProbe: {
              httpGet: {
                path: '/health?ready=1',
                port: 'http',
                scheme: 'HTTPS',
              },
              initialDelaySeconds: 10,
              periodSeconds: 5,
              failureThreshold: 2,
            },
            volumeMounts: [
              {
                name: 'datadir',
                mountPath: '/cockroach/cockroach-data',
              },
              {
                name: 'certs',
                mountPath: '/cockroach/cockroach-certs',
              },
            ],
            env: [
              {
                name: 'COCKROACH_CHANNEL',
                value: {
                  '[object Object]': null,
                },
              },
            ],
            command: [
              '/bin/bash',
              '-ecx',
              'exec /cockroach/cockroach start --logtostderr --certs-dir "/cockroach/cockroach-certs" --advertise-addr "$(printenv EXTERNAL_HOSTNAME_FOR_$HOSTNAME)" --locality-advertise-addr "zone={{ .Values.Locality }}@$(hostname -f)" --http-addr 0.0.0.0 --join "{{ range $index := list 0 1 2 -}}\n    {{- if ne $index 0 }},{{ end -}}\n    cockroachdb-{{ $index }}.cockroachdb.{{ $.Values.Namespace }}.svc.cluster.local:{{ $.Values.CockroachPort }}\n  {{- end -}}\n  {{- if .Values.JoinExisting -}}\n    ,{{ join "," .Values.JoinExisting -}}\n  {{ end }}"\n--locality "zone={{ .Values.Locality }}" --cache 25% --max-sql-memory 25%',
            ],
          },
        ],
        terminationGracePeriodSeconds: 60,
        volumes: [
          {
            name: 'datadir',
            persistentVolumeClaim: {
              claimName: 'datadir',
            },
          },
          {
            name: 'certs',
            secret: {
              secretName: 'cockroachdb.node',
              defaultMode: 256,
            },
          },
        ],
      },
    },
    podManagementPolicy: 'Parallel',
    updateStrategy: {
      type: 'RollingUpdate',
    },
    volumeClaimTemplates: [
      {
        metadata: {
          name: 'datadir',
        },
        spec: {
          storageClassName: {
            '[object Object]': null,
          },
          accessModes: [
            'ReadWriteOnce',
          ],
          resources: {
            requests: {
              storage: {
                '[object Object]': null,
              },
            },
          },
        },
      },
    ],
  },
}