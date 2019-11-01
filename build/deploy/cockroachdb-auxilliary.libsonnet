{
  apiVersion: 'v1',
  kind: 'ServiceAccount',
  metadata: {
    name: 'cockroachdb',
    labels: {
      app: 'cockroachdb',
    },
  },
}


{
  apiVersion: 'rbac.authorization.k8s.io/v1beta1',
  kind: 'Role',
  metadata: {
    name: 'cockroachdb',
    labels: {
      app: 'cockroachdb',
    },
  },
  rules: [
    {
      apiGroups: [
        '',
      ],
      resources: [
        'secrets',
      ],
      verbs: [
        'create',
        'get',
      ],
    },
  ],
}


{
  apiVersion: 'rbac.authorization.k8s.io/v1beta1',
  kind: 'ClusterRole',
  metadata: {
    name: 'cockroachdb',
    labels: {
      app: 'cockroachdb',
    },
  },
  rules: [
    {
      apiGroups: [
        'certificates.k8s.io',
      ],
      resources: [
        'certificatesigningrequests',
      ],
      verbs: [
        'create',
        'get',
        'watch',
      ],
    },
  ],
}

{
  apiVersion: 'rbac.authorization.k8s.io/v1beta1',
  kind: 'RoleBinding',
  metadata: {
    name: 'cockroachdb',
    labels: {
      app: 'cockroachdb',
    },
  },
  roleRef: {
    apiGroup: 'rbac.authorization.k8s.io',
    kind: 'Role',
    name: 'cockroachdb',
  },
  subjects: [
    {
      kind: 'ServiceAccount',
      name: 'cockroachdb',
    },
  ],
}

{
  apiVersion: 'rbac.authorization.k8s.io/v1beta1',
  kind: 'ClusterRoleBinding',
  metadata: {
    name: 'cockroachdb',
    labels: {
      app: 'cockroachdb',
    },
  },
  roleRef: {
    apiGroup: 'rbac.authorization.k8s.io',
    kind: 'ClusterRole',
    name: 'cockroachdb',
  },
  subjects: [
    {
      kind: 'ServiceAccount',
      name: 'cockroachdb',
      namespace: 'default',
    },
  ],
}



{
  apiVersion: 'v1',
  kind: 'Service',
  metadata: {
    name: 'cockroachdb',
    labels: {
      app: 'cockroachdb',
    },
    annotations: {
      'service.alpha.kubernetes.io/tolerate-unready-endpoints': 'true',
      'prometheus.io/scrape': 'true',
      'prometheus.io/path': '_status/vars',
      'prometheus.io/port': '{{ .Values.HttpPort }}',
    },
  },
  spec: {
    ports: [
      {
        port: {
          '[object Object]': null,
        },
        targetPort: {
          '[object Object]': null,
        },
        name: 'cockroach',
      },
      {
        port: {
          '[object Object]': null,
        },
        targetPort: {
          '[object Object]': null,
        },
        name: 'http',
      },
    ],
    publishNotReadyAddresses: true,
    clusterIP: 'None',
    selector: {
      app: 'cockroachdb',
    },
  },
}


{
  apiVersion: 'policy/v1beta1',
  kind: 'PodDisruptionBudget',
  metadata: {
    name: 'cockroachdb-budget',
    labels: {
      app: 'cockroachdb',
    },
  },
  spec: {
    selector: {
      matchLabels: {
        app: 'cockroachdb',
      },
    },
    maxUnavailable: 1,
  },
}


{
  apiVersion: 'v1',
  kind: 'Service',
  metadata: {
    name: 'cockroachdb-balanced',
    labels: {
      app: 'cockroachdb',
    },
  },
  spec: {
    ports: [
      {
        port: {
          '[object Object]': null,
        },
        targetPort: {
          '[object Object]': null,
        },
        name: 'cockroach',
      },
      {
        port: {
          '[object Object]': null,
        },
        targetPort: {
          '[object Object]': null,
        },
        name: 'http',
      },
    ],
    selector: {
      app: 'cockroachdb',
    },
    sessionAffinity: 'ClientIP',
  },
}