{
  "istio-obj-0": {
    "apiVersion": "rbac.authorization.k8s.io/v1",
    "kind": "ClusterRole",
    "metadata": {
      "name": "istio-reader-istio-system",
      "labels": {
        "app": "istio-reader",
        "release": "istio"
      }
    },
    "rules": [
      {
        "apiGroups": [
          "config.istio.io",
          "rbac.istio.io",
          "security.istio.io",
          "networking.istio.io",
          "authentication.istio.io"
        ],
        "resources": [
          "*"
        ],
        "verbs": [
          "get",
          "list",
          "watch"
        ]
      },
      {
        "apiGroups": [
          ""
        ],
        "resources": [
          "endpoints",
          "pods",
          "services",
          "nodes",
          "replicationcontrollers"
        ],
        "verbs": [
          "get",
          "list",
          "watch"
        ]
      },
      {
        "apiGroups": [
          "apps"
        ],
        "resources": [
          "replicasets"
        ],
        "verbs": [
          "get",
          "list",
          "watch"
        ]
      }
    ]
  },
  "istio-obj-1": {
    "apiVersion": "rbac.authorization.k8s.io/v1",
    "kind": "ClusterRoleBinding",
    "metadata": {
      "name": "istio-reader-istio-system",
      "labels": {
        "app": "istio-reader",
        "release": "istio"
      }
    },
    "roleRef": {
      "apiGroup": "rbac.authorization.k8s.io",
      "kind": "ClusterRole",
      "name": "istio-reader-istio-system"
    },
    "subjects": [
      {
        "kind": "ServiceAccount",
        "name": "istio-reader-service-account",
        "namespace": "istio-system"
      }
    ]
  },
  "istio-obj-25": {
    "apiVersion": "v1",
    "kind": "Namespace",
    "metadata": {
      "name": "istio-system",
      "labels": {
        "istio-operator-managed": "Reconcile",
        "istio-injection": "disabled"
      }
    }
  },
  "istio-obj-26": {
    "apiVersion": "v1",
    "kind": "ServiceAccount",
    "metadata": {
      "name": "istio-reader-service-account",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-reader",
        "release": "istio"
      }
    }
  },
  "istio-obj-27": {
    "apiVersion": "rbac.authorization.k8s.io/v1",
    "kind": "ClusterRole",
    "metadata": {
      "name": "istio-citadel-istio-system",
      "labels": {
        "app": "citadel",
        "release": "istio"
      }
    },
    "rules": [
      {
        "apiGroups": [
          ""
        ],
        "resources": [
          "configmaps"
        ],
        "verbs": [
          "create",
          "get",
          "update"
        ]
      },
      {
        "apiGroups": [
          ""
        ],
        "resources": [
          "secrets"
        ],
        "verbs": [
          "create",
          "get",
          "watch",
          "list",
          "update",
          "delete"
        ]
      },
      {
        "apiGroups": [
          ""
        ],
        "resources": [
          "serviceaccounts",
          "services",
          "namespaces"
        ],
        "verbs": [
          "get",
          "watch",
          "list"
        ]
      },
      {
        "apiGroups": [
          "authentication.k8s.io"
        ],
        "resources": [
          "tokenreviews"
        ],
        "verbs": [
          "create"
        ]
      }
    ]
  },
  "istio-obj-28": {
    "apiVersion": "rbac.authorization.k8s.io/v1",
    "kind": "ClusterRoleBinding",
    "metadata": {
      "name": "istio-citadel-istio-system",
      "labels": {
        "release": "istio"
      }
    },
    "roleRef": {
      "apiGroup": "rbac.authorization.k8s.io",
      "kind": "ClusterRole",
      "name": "istio-citadel-istio-system"
    },
    "subjects": [
      {
        "kind": "ServiceAccount",
        "name": "istio-citadel-service-account",
        "namespace": "istio-system"
      }
    ]
  },
  "istio-obj-29": {
    "apiVersion": "apps/v1",
    "kind": "Deployment",
    "metadata": {
      "labels": {
        "app": "security",
        "istio": "citadel",
        "release": "istio"
      },
      "name": "istio-citadel",
      "namespace": "istio-system"
    },
    "spec": {
      "replicas": 1,
      "selector": {
        "matchLabels": {
          "istio": "citadel"
        }
      },
      "strategy": {
        "rollingUpdate": {
          "maxSurge": "100%",
          "maxUnavailable": "25%"
        }
      },
      "template": {
        "metadata": {
          "annotations": {
            "sidecar.istio.io/inject": "false"
          },
          "labels": {
            "app": "citadel",
            "istio": "citadel"
          }
        },
        "spec": {
          "affinity": {
            "nodeAffinity": {
              "preferredDuringSchedulingIgnoredDuringExecution": [
                {
                  "preference": {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "amd64"
                        ]
                      }
                    ]
                  },
                  "weight": 2
                },
                {
                  "preference": {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "ppc64le"
                        ]
                      }
                    ]
                  },
                  "weight": 2
                },
                {
                  "preference": {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "s390x"
                        ]
                      }
                    ]
                  },
                  "weight": 2
                }
              ],
              "requiredDuringSchedulingIgnoredDuringExecution": {
                "nodeSelectorTerms": [
                  {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "amd64",
                          "ppc64le",
                          "s390x"
                        ]
                      }
                    ]
                  }
                ]
              }
            }
          },
          "containers": [
            {
              "args": [
                "--append-dns-names=true",
                "--grpc-port=8060",
                "--citadel-storage-namespace=istio-system",
                "--custom-dns-names=istio-galley-service-account.istio-config:istio-galley.istio-config.svc,istio-galley-service-account.istio-control:istio-galley.istio-control.svc,istio-galley-service-account.istio-control-master:istio-galley.istio-control-master.svc,istio-galley-service-account.istio-master:istio-galley.istio-master.svc,istio-galley-service-account.istio-pilot11:istio-galley.istio-pilot11.svc,istio-pilot-service-account.istio-control:istio-pilot.istio-control,istio-pilot-service-account.istio-pilot11:istio-pilot.istio-system,istio-sidecar-injector-service-account.istio-control:istio-sidecar-injector.istio-control.svc,istio-sidecar-injector-service-account.istio-control-master:istio-sidecar-injector.istio-control-master.svc,istio-sidecar-injector-service-account.istio-master:istio-sidecar-injector.istio-master.svc,istio-sidecar-injector-service-account.istio-pilot11:istio-sidecar-injector.istio-pilot11.svc,istio-sidecar-injector-service-account.istio-remote:istio-sidecar-injector.istio-remote.svc,",
                "--self-signed-ca=true",
                "--trust-domain=cluster.local",
                "--workload-cert-ttl=2160h"
              ],
              "env": [
                {
                  "name": "CITADEL_ENABLE_NAMESPACES_BY_DEFAULT",
                  "value": "true"
                }
              ],
              "image": "docker.io/istio/citadel:1.4.3",
              "imagePullPolicy": "IfNotPresent",
              "livenessProbe": {
                "httpGet": {
                  "path": "/version",
                  "port": 15014
                },
                "initialDelaySeconds": 5,
                "periodSeconds": 5
              },
              "name": "citadel",
              "resources": {
                "requests": {
                  "cpu": "10m"
                }
              }
            }
          ],
          "serviceAccountName": "istio-citadel-service-account"
        }
      }
    }
  },
  "istio-obj-30": {
    "apiVersion": "policy/v1beta1",
    "kind": "PodDisruptionBudget",
    "metadata": {
      "name": "istio-citadel",
      "namespace": "istio-system",
      "labels": {
        "app": "security",
        "istio": "citadel",
        "release": "istio"
      }
    },
    "spec": {
      "minAvailable": 1,
      "selector": {
        "matchLabels": {
          "app": "citadel",
          "istio": "citadel"
        }
      }
    }
  },
  "istio-obj-31": {
    "apiVersion": "v1",
    "kind": "Service",
    "metadata": {
      "name": "istio-citadel",
      "namespace": "istio-system",
      "labels": {
        "app": "security",
        "istio": "citadel",
        "release": "istio"
      }
    },
    "spec": {
      "ports": [
        {
          "name": "grpc-citadel",
          "port": 8060,
          "targetPort": 8060,
          "protocol": "TCP"
        },
        {
          "name": "http-monitoring",
          "port": 15014
        }
      ],
      "selector": {
        "app": "citadel"
      }
    }
  },
  "istio-obj-32": {
    "apiVersion": "v1",
    "kind": "ServiceAccount",
    "metadata": {
      "name": "istio-citadel-service-account",
      "namespace": "istio-system",
      "labels": {
        "app": "security",
        "release": "istio"
      }
    }
  },
  "istio-obj-33": {
    "apiVersion": "rbac.authorization.k8s.io/v1",
    "kind": "ClusterRole",
    "metadata": {
      "name": "istio-galley-istio-system",
      "labels": {
        "release": "istio"
      }
    },
    "rules": [
      {
        "apiGroups": [
          "authentication.istio.io",
          "config.istio.io",
          "networking.istio.io",
          "rbac.istio.io",
          "security.istio.io"
        ],
        "resources": [
          "*"
        ],
        "verbs": [
          "get",
          "list",
          "watch"
        ]
      },
      {
        "apiGroups": [
          "authentication.istio.io",
          "config.istio.io",
          "networking.istio.io",
          "rbac.istio.io",
          "security.istio.io"
        ],
        "resources": [
          "*/status"
        ],
        "verbs": [
          "update"
        ]
      },
      {
        "apiGroups": [
          "admissionregistration.k8s.io"
        ],
        "resources": [
          "validatingwebhookconfigurations"
        ],
        "verbs": [
          "*"
        ]
      },
      {
        "apiGroups": [
          "extensions",
          "apps"
        ],
        "resources": [
          "deployments"
        ],
        "resourceNames": [
          "istio-galley"
        ],
        "verbs": [
          "get"
        ]
      },
      {
        "apiGroups": [
          ""
        ],
        "resources": [
          "pods",
          "nodes",
          "services",
          "endpoints",
          "namespaces"
        ],
        "verbs": [
          "get",
          "list",
          "watch"
        ]
      },
      {
        "apiGroups": [
          "extensions"
        ],
        "resources": [
          "ingresses"
        ],
        "verbs": [
          "get",
          "list",
          "watch"
        ]
      },
      {
        "apiGroups": [
          "extensions"
        ],
        "resources": [
          "deployments/finalizers"
        ],
        "resourceNames": [
          "istio-galley"
        ],
        "verbs": [
          "update"
        ]
      },
      {
        "apiGroups": [
          "apiextensions.k8s.io"
        ],
        "resources": [
          "customresourcedefinitions"
        ],
        "verbs": [
          "get",
          "list",
          "watch"
        ]
      },
      {
        "apiGroups": [
          "rbac.authorization.k8s.io"
        ],
        "resources": [
          "clusterroles"
        ],
        "verbs": [
          "get",
          "list",
          "watch"
        ]
      }
    ]
  },
  "istio-obj-34": {
    "apiVersion": "rbac.authorization.k8s.io/v1",
    "kind": "ClusterRoleBinding",
    "metadata": {
      "name": "istio-galley-admin-role-binding-istio-system",
      "labels": {
        "release": "istio"
      }
    },
    "roleRef": {
      "apiGroup": "rbac.authorization.k8s.io",
      "kind": "ClusterRole",
      "name": "istio-galley-istio-system"
    },
    "subjects": [
      {
        "kind": "ServiceAccount",
        "name": "istio-galley-service-account",
        "namespace": "istio-system"
      }
    ]
  },
  "istio-obj-35": {
    "apiVersion": "v1",
    "kind": "ConfigMap",
    "metadata": {
      "namespace": "istio-system",
      "name": "galley-envoy-config",
      "labels": {
        "app": "galley",
        "istio": "galley",
        "release": "istio"
      }
    },
    "data": {
      "envoy.yaml.tmpl": "admin:\n  access_log_path: /dev/null\n  address:\n    socket_address:\n      address: 127.0.0.1\n      port_value: 15000\n\nstatic_resources:\n\n  clusters:\n  - name: in.9901\n    http2_protocol_options: {}\n    connect_timeout: 1.000s\n\n    hosts:\n    - socket_address:\n        address: 127.0.0.1\n        port_value: 9901\n\n    circuit_breakers:\n      thresholds:\n      - max_connections: 100000\n        max_pending_requests: 100000\n        max_requests: 100000\n        max_retries: 3\n\n  listeners:\n  - name: \"15019\"\n    address:\n      socket_address:\n        address: 0.0.0.0\n        port_value: 15019\n    filter_chains:\n    - filters:\n      - name: envoy.http_connection_manager\n        config:\n          codec_type: HTTP2\n          stat_prefix: \"15010\"\n          http2_protocol_options:\n            max_concurrent_streams: 1073741824\n\n          access_log:\n          - name: envoy.file_access_log\n            config:\n              path: /dev/stdout\n\n          http_filters:\n          - name: envoy.router\n\n          route_config:\n            name: \"15019\"\n\n            virtual_hosts:\n            - name: istio-galley\n\n              domains:\n              - '*'\n\n              routes:\n              - match:\n                  prefix: /\n                route:\n                  cluster: in.9901\n                  timeout: 0.000s\n      tls_context:\n        common_tls_context:\n          alpn_protocols:\n          - h2\n          tls_certificates:\n          - certificate_chain:\n              filename: /etc/certs/cert-chain.pem\n            private_key:\n              filename: /etc/certs/key.pem\n          validation_context:\n            trusted_ca:\n              filename: /etc/certs/root-cert.pem\n        require_client_certificate: true"
    }
  },
  "istio-obj-36": {
    "apiVersion": "v1",
    "kind": "ConfigMap",
    "metadata": {
      "name": "istio-mesh-galley",
      "namespace": "istio-system",
      "labels": {
        "release": "istio"
      }
    },
    "data": {
      "mesh": "{}"
    }
  },
  "istio-obj-37": {
    "apiVersion": "v1",
    "kind": "ConfigMap",
    "metadata": {
      "name": "istio-galley-configuration",
      "namespace": "istio-system",
      "labels": {
        "release": "istio"
      }
    },
    "data": {
      "validatingwebhookconfiguration.yaml": "apiVersion: admissionregistration.k8s.io/v1beta1\nkind: ValidatingWebhookConfiguration\nmetadata:\n  name: istio-galley-istio-system\n  namespace: istio-system\n  labels:\n    app: galley\n    release: istio\n    istio: galley\nwebhooks:\n  - name: pilot.validation.istio.io\n    clientConfig:\n      service:\n        name: istio-galley\n        namespace: istio-system\n        path: \"/admitpilot\"\n      caBundle: \"\"\n    rules:\n      - operations:\n        - CREATE\n        - UPDATE\n        apiGroups:\n        - config.istio.io\n        apiVersions:\n        - v1alpha2\n        resources:\n        - httpapispecs\n        - httpapispecbindings\n        - quotaspecs\n        - quotaspecbindings\n      - operations:\n        - CREATE\n        - UPDATE\n        apiGroups:\n        - rbac.istio.io\n        apiVersions:\n        - \"*\"\n        resources:\n        - \"*\"\n      - operations:\n        - CREATE\n        - UPDATE\n        apiGroups:\n        - security.istio.io\n        apiVersions:\n        - \"*\"\n        resources:\n        - \"*\"\n      - operations:\n        - CREATE\n        - UPDATE\n        apiGroups:\n        - authentication.istio.io\n        apiVersions:\n        - \"*\"\n        resources:\n        - \"*\"\n      - operations:\n        - CREATE\n        - UPDATE\n        apiGroups:\n        - networking.istio.io\n        apiVersions:\n        - \"*\"\n        resources:\n        - destinationrules\n        - envoyfilters\n        - gateways\n        - serviceentries\n        - sidecars\n        - virtualservices\n    failurePolicy: Fail\n    sideEffects: None\n  - name: mixer.validation.istio.io\n    clientConfig:\n      service:\n        name: istio-galley\n        namespace: istio-system\n        path: \"/admitmixer\"\n      caBundle: \"\"\n    rules:\n      - operations:\n        - CREATE\n        - UPDATE\n        apiGroups:\n        - config.istio.io\n        apiVersions:\n        - v1alpha2\n        resources:\n        - rules\n        - attributemanifests\n        - circonuses\n        - deniers\n        - fluentds\n        - kubernetesenvs\n        - listcheckers\n        - memquotas\n        - noops\n        - opas\n        - prometheuses\n        - rbacs\n        - solarwindses\n        - stackdrivers\n        - cloudwatches\n        - dogstatsds\n        - statsds\n        - stdios\n        - apikeys\n        - authorizations\n        - checknothings\n        # - kuberneteses\n        - listentries\n        - logentries\n        - metrics\n        - quotas\n        - reportnothings\n        - tracespans\n        - adapters\n        - handlers\n        - instances\n        - templates\n        - zipkins\n    failurePolicy: Fail\n    sideEffects: None"
    }
  },
  "istio-obj-38": {
    "apiVersion": "apps/v1",
    "kind": "Deployment",
    "metadata": {
      "labels": {
        "app": "galley",
        "istio": "galley",
        "release": "istio"
      },
      "name": "istio-galley",
      "namespace": "istio-system"
    },
    "spec": {
      "replicas": 1,
      "selector": {
        "matchLabels": {
          "istio": "galley"
        }
      },
      "strategy": {
        "rollingUpdate": {
          "maxSurge": "100%",
          "maxUnavailable": "25%"
        }
      },
      "template": {
        "metadata": {
          "annotations": {
            "sidecar.istio.io/inject": "false"
          },
          "labels": {
            "app": "galley",
            "chart": "galley",
            "heritage": "Tiller",
            "istio": "galley",
            "release": "istio"
          }
        },
        "spec": {
          "affinity": {
            "nodeAffinity": {
              "preferredDuringSchedulingIgnoredDuringExecution": [
                {
                  "preference": {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "amd64"
                        ]
                      }
                    ]
                  },
                  "weight": 2
                },
                {
                  "preference": {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "ppc64le"
                        ]
                      }
                    ]
                  },
                  "weight": 2
                },
                {
                  "preference": {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "s390x"
                        ]
                      }
                    ]
                  },
                  "weight": 2
                }
              ],
              "requiredDuringSchedulingIgnoredDuringExecution": {
                "nodeSelectorTerms": [
                  {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "amd64",
                          "ppc64le",
                          "s390x"
                        ]
                      }
                    ]
                  }
                ]
              }
            }
          },
          "containers": [
            {
              "command": [
                "/usr/local/bin/galley",
                "server",
                "--meshConfigFile=/etc/mesh-config/mesh",
                "--livenessProbeInterval=1s",
                "--livenessProbePath=/tmp/healthliveness",
                "--readinessProbePath=/tmp/healthready",
                "--readinessProbeInterval=1s",
                "--insecure=true",
                "--enable-validation=true",
                "--enable-reconcileWebhookConfiguration=true",
                "--enable-server=true",
                "--deployment-namespace=istio-system",
                "--validation-webhook-config-file",
                "/etc/config/validatingwebhookconfiguration.yaml",
                "--monitoringPort=15014",
                "--validation-port=9443",
                "--log_output_level=default:info"
              ],
              "image": "docker.io/istio/galley:1.4.3",
              "imagePullPolicy": "IfNotPresent",
              "livenessProbe": {
                "exec": {
                  "command": [
                    "/usr/local/bin/galley",
                    "probe",
                    "--probe-path=/tmp/healthliveness",
                    "--interval=10s"
                  ]
                },
                "initialDelaySeconds": 5,
                "periodSeconds": 5
              },
              "name": "galley",
              "ports": [
                {
                  "containerPort": 9443
                },
                {
                  "containerPort": 15014
                },
                {
                  "containerPort": 15019
                },
                {
                  "containerPort": 9901
                }
              ],
              "readinessProbe": {
                "exec": {
                  "command": [
                    "/usr/local/bin/galley",
                    "probe",
                    "--probe-path=/tmp/healthready",
                    "--interval=10s"
                  ]
                },
                "initialDelaySeconds": 5,
                "periodSeconds": 5
              },
              "resources": {
                "requests": {
                  "cpu": "100m"
                }
              },
              "volumeMounts": [
                {
                  "mountPath": "/etc/certs",
                  "name": "istio-certs",
                  "readOnly": true
                },
                {
                  "mountPath": "/etc/config",
                  "name": "config",
                  "readOnly": true
                },
                {
                  "mountPath": "/etc/mesh-config",
                  "name": "mesh-config",
                  "readOnly": true
                }
              ]
            },
            {
              "args": [
                "proxy",
                "--serviceCluster",
                "istio-galley",
                "--templateFile",
                "/var/lib/istio/galley/envoy/envoy.yaml.tmpl",
                "--controlPlaneAuthPolicy",
                "MUTUAL_TLS",
                "--trust-domain=cluster.local"
              ],
              "env": [
                {
                  "name": "POD_NAME",
                  "valueFrom": {
                    "fieldRef": {
                      "apiVersion": "v1",
                      "fieldPath": "metadata.name"
                    }
                  }
                },
                {
                  "name": "POD_NAMESPACE",
                  "valueFrom": {
                    "fieldRef": {
                      "apiVersion": "v1",
                      "fieldPath": "metadata.namespace"
                    }
                  }
                },
                {
                  "name": "INSTANCE_IP",
                  "valueFrom": {
                    "fieldRef": {
                      "apiVersion": "v1",
                      "fieldPath": "status.podIP"
                    }
                  }
                },
                {
                  "name": "SDS_ENABLED",
                  "value": "false"
                }
              ],
              "image": "docker.io/istio/proxyv2:1.4.3",
              "imagePullPolicy": "IfNotPresent",
              "name": "istio-proxy",
              "ports": [
                {
                  "containerPort": 9902
                }
              ],
              "resources": {
                "limits": {
                  "cpu": "2000m",
                  "memory": "1024Mi"
                },
                "requests": {
                  "cpu": "100m",
                  "memory": "128Mi"
                }
              },
              "volumeMounts": [
                {
                  "mountPath": "/var/lib/istio/galley/envoy",
                  "name": "envoy-config"
                },
                {
                  "mountPath": "/etc/certs",
                  "name": "istio-certs",
                  "readOnly": true
                }
              ]
            }
          ],
          "serviceAccountName": "istio-galley-service-account",
          "volumes": [
            {
              "name": "istio-certs",
              "secret": {
                "secretName": "istio.istio-galley-service-account"
              }
            },
            {
              "configMap": {
                "name": "galley-envoy-config"
              },
              "name": "envoy-config"
            },
            {
              "configMap": {
                "name": "istio-galley-configuration"
              },
              "name": "config"
            },
            {
              "configMap": {
                "name": "istio-mesh-galley"
              },
              "name": "mesh-config"
            }
          ]
        }
      }
    }
  },
  "istio-obj-39": {
    "apiVersion": "policy/v1beta1",
    "kind": "PodDisruptionBudget",
    "metadata": {
      "name": "istio-galley",
      "namespace": "istio-system",
      "labels": {
        "app": "galley",
        "release": "istio",
        "istio": "galley"
      }
    },
    "spec": {
      "minAvailable": 1,
      "selector": {
        "matchLabels": {
          "app": "galley",
          "release": "istio",
          "istio": "galley"
        }
      }
    }
  },
  "istio-obj-40": {
    "apiVersion": "v1",
    "kind": "Service",
    "metadata": {
      "name": "istio-galley",
      "namespace": "istio-system",
      "labels": {
        "app": "galley",
        "istio": "galley",
        "release": "istio"
      }
    },
    "spec": {
      "ports": [
        {
          "port": 443,
          "name": "https-validation",
          "targetPort": 9443
        },
        {
          "port": 15014,
          "name": "http-monitoring"
        },
        {
          "port": 9901,
          "name": "grpc-mcp"
        },
        {
          "port": 15019,
          "name": "grpc-tls-mcp"
        }
      ],
      "selector": {
        "istio": "galley"
      }
    }
  },
  "istio-obj-41": {
    "apiVersion": "v1",
    "kind": "ServiceAccount",
    "metadata": {
      "name": "istio-galley-service-account",
      "namespace": "istio-system",
      "labels": {
        "app": "galley",
        "release": "istio"
      }
    }
  },
  "istio-obj-42": {
    "apiVersion": "autoscaling/v2beta1",
    "kind": "HorizontalPodAutoscaler",
    "metadata": {
      "labels": {
        "app": "istio-ingressgateway",
        "istio": "ingressgateway",
        "release": "istio"
      },
      "name": "istio-ingressgateway",
      "namespace": "istio-system"
    },
    "spec": {
      "maxReplicas": 5,
      "metrics": [
        {
          "resource": {
            "name": "cpu",
            "targetAverageUtilization": 80
          },
          "type": "Resource"
        }
      ],
      "minReplicas": 1,
      "scaleTargetRef": {
        "apiVersion": "apps/v1",
        "kind": "Deployment",
        "name": "istio-ingressgateway"
      }
    }
  },
  "istio-obj-43": {
    "apiVersion": "apps/v1",
    "kind": "Deployment",
    "metadata": {
      "labels": {
        "app": "istio-ingressgateway",
        "istio": "ingressgateway",
        "release": "istio"
      },
      "name": "istio-ingressgateway",
      "namespace": "istio-system"
    },
    "spec": {
      "selector": {
        "matchLabels": {
          "app": "istio-ingressgateway",
          "istio": "ingressgateway"
        }
      },
      "strategy": {
        "rollingUpdate": {
          "maxSurge": "100%",
          "maxUnavailable": "25%"
        }
      },
      "template": {
        "metadata": {
          "annotations": {
            "sidecar.istio.io/inject": "false"
          },
          "labels": {
            "app": "istio-ingressgateway",
            "chart": "gateways",
            "heritage": "Tiller",
            "istio": "ingressgateway",
            "release": "istio"
          }
        },
        "spec": {
          "affinity": {
            "nodeAffinity": {
              "preferredDuringSchedulingIgnoredDuringExecution": [
                {
                  "preference": {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "amd64"
                        ]
                      }
                    ]
                  },
                  "weight": 2
                },
                {
                  "preference": {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "ppc64le"
                        ]
                      }
                    ]
                  },
                  "weight": 2
                },
                {
                  "preference": {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "s390x"
                        ]
                      }
                    ]
                  },
                  "weight": 2
                }
              ],
              "requiredDuringSchedulingIgnoredDuringExecution": {
                "nodeSelectorTerms": [
                  {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "amd64",
                          "ppc64le",
                          "s390x"
                        ]
                      }
                    ]
                  }
                ]
              }
            }
          },
          "containers": [
            {
              "env": [
                {
                  "name": "ENABLE_WORKLOAD_SDS",
                  "value": "false"
                },
                {
                  "name": "ENABLE_INGRESS_GATEWAY_SDS",
                  "value": "true"
                },
                {
                  "name": "INGRESS_GATEWAY_NAMESPACE",
                  "valueFrom": {
                    "fieldRef": {
                      "apiVersion": "v1",
                      "fieldPath": "metadata.namespace"
                    }
                  }
                }
              ],
              "image": "docker.io/istio/node-agent-k8s:1.4.3",
              "imagePullPolicy": "IfNotPresent",
              "name": "ingress-sds",
              "resources": {
                "limits": {
                  "cpu": "2000m",
                  "memory": "1024Mi"
                },
                "requests": {
                  "cpu": "100m",
                  "memory": "128Mi"
                }
              },
              "volumeMounts": [
                {
                  "mountPath": "/var/run/ingress_gateway",
                  "name": "ingressgatewaysdsudspath"
                }
              ]
            },
            {
              "args": [
                "proxy",
                "router",
                "--domain",
                "$(POD_NAMESPACE).svc.cluster.local",
                "--proxyLogLevel=warning",
                "--proxyComponentLogLevel=misc:error",
                "--log_output_level=default:info",
                "--drainDuration",
                "45s",
                "--parentShutdownDuration",
                "1m0s",
                "--connectTimeout",
                "10s",
                "--serviceCluster",
                "istio-ingressgateway",
                "--zipkinAddress",
                "zipkin.istio-system:9411",
                "--proxyAdminPort",
                "15000",
                "--statusPort",
                "15020",
                "--controlPlaneAuthPolicy",
                "MUTUAL_TLS",
                "--discoveryAddress",
                "istio-pilot.istio-system:15011",
                "--trust-domain=cluster.local"
              ],
              "env": [
                {
                  "name": "NODE_NAME",
                  "valueFrom": {
                    "fieldRef": {
                      "apiVersion": "v1",
                      "fieldPath": "spec.nodeName"
                    }
                  }
                },
                {
                  "name": "POD_NAME",
                  "valueFrom": {
                    "fieldRef": {
                      "apiVersion": "v1",
                      "fieldPath": "metadata.name"
                    }
                  }
                },
                {
                  "name": "POD_NAMESPACE",
                  "valueFrom": {
                    "fieldRef": {
                      "apiVersion": "v1",
                      "fieldPath": "metadata.namespace"
                    }
                  }
                },
                {
                  "name": "INSTANCE_IP",
                  "valueFrom": {
                    "fieldRef": {
                      "apiVersion": "v1",
                      "fieldPath": "status.podIP"
                    }
                  }
                },
                {
                  "name": "HOST_IP",
                  "valueFrom": {
                    "fieldRef": {
                      "apiVersion": "v1",
                      "fieldPath": "status.hostIP"
                    }
                  }
                },
                {
                  "name": "SERVICE_ACCOUNT",
                  "valueFrom": {
                    "fieldRef": {
                      "fieldPath": "spec.serviceAccountName"
                    }
                  }
                },
                {
                  "name": "ISTIO_META_WORKLOAD_NAME",
                  "value": "istio-ingressgateway"
                },
                {
                  "name": "ISTIO_META_OWNER",
                  "value": "kubernetes://apis/apps/v1/namespaces/istio-system/deployments/istio-ingressgateway"
                },
                {
                  "name": "ISTIO_META_MESH_ID",
                  "value": "cluster.local"
                },
                {
                  "name": "ISTIO_META_POD_NAME",
                  "valueFrom": {
                    "fieldRef": {
                      "apiVersion": "v1",
                      "fieldPath": "metadata.name"
                    }
                  }
                },
                {
                  "name": "ISTIO_META_CONFIG_NAMESPACE",
                  "valueFrom": {
                    "fieldRef": {
                      "fieldPath": "metadata.namespace"
                    }
                  }
                },
                {
                  "name": "ISTIO_META_USER_SDS",
                  "value": "true"
                },
                {
                  "name": "ISTIO_META_ROUTER_MODE",
                  "value": "sni-dnat"
                },
                {
                  "name": "ISTIO_METAJSON_LABELS",
                  "value": "{\"app\":\"istio-ingressgateway\",\"istio\":\"ingressgateway\"}\n"
                },
                {
                  "name": "ISTIO_META_CLUSTER_ID",
                  "value": "Kubernetes"
                },
                {
                  "name": "SDS_ENABLED",
                  "value": "false"
                }
              ],
              "image": "docker.io/istio/proxyv2:1.4.3",
              "imagePullPolicy": "IfNotPresent",
              "name": "istio-proxy",
              "ports": [
                {
                  "containerPort": 15020
                },
                {
                  "containerPort": 80
                },
                {
                  "containerPort": 443
                },
                {
                  "containerPort": 15029
                },
                {
                  "containerPort": 15030
                },
                {
                  "containerPort": 15031
                },
                {
                  "containerPort": 15032
                },
                {
                  "containerPort": 15443
                },
                {
                  "containerPort": 15011
                },
                {
                  "containerPort": 8060
                },
                {
                  "containerPort": 853
                },
                {
                  "containerPort": 15090,
                  "name": "http-envoy-prom",
                  "protocol": "TCP"
                }
              ],
              "readinessProbe": {
                "failureThreshold": 30,
                "httpGet": {
                  "path": "/healthz/ready",
                  "port": 15020,
                  "scheme": "HTTP"
                },
                "initialDelaySeconds": 1,
                "periodSeconds": 2,
                "successThreshold": 1,
                "timeoutSeconds": 1
              },
              "resources": {
                "limits": {
                  "cpu": "2000m",
                  "memory": "1024Mi"
                },
                "requests": {
                  "cpu": "100m",
                  "memory": "128Mi"
                }
              },
              "volumeMounts": [
                {
                  "mountPath": "/etc/certs",
                  "name": "istio-certs",
                  "readOnly": true
                },
                {
                  "mountPath": "/var/run/ingress_gateway",
                  "name": "ingressgatewaysdsudspath"
                },
                {
                  "mountPath": "/etc/istio/ingressgateway-certs",
                  "name": "ingressgateway-certs",
                  "readOnly": true
                },
                {
                  "mountPath": "/etc/istio/ingressgateway-ca-certs",
                  "name": "ingressgateway-ca-certs",
                  "readOnly": true
                }
              ]
            }
          ],
          "serviceAccountName": "istio-ingressgateway-service-account",
          "volumes": [
            {
              "emptyDir": {},
              "name": "ingressgatewaysdsudspath"
            },
            {
              "name": "istio-certs",
              "secret": {
                "optional": true,
                "secretName": "istio.istio-ingressgateway-service-account"
              }
            },
            {
              "name": "ingressgateway-certs",
              "secret": {
                "optional": true,
                "secretName": "istio-ingressgateway-certs"
              }
            },
            {
              "name": "ingressgateway-ca-certs",
              "secret": {
                "optional": true,
                "secretName": "istio-ingressgateway-ca-certs"
              }
            }
          ]
        }
      }
    }
  },
  "istio-obj-44": {
    "apiVersion": "networking.istio.io/v1alpha3",
    "kind": "Gateway",
    "metadata": {
      "name": "ingressgateway",
      "namespace": "istio-system",
      "labels": {
        "release": "istio"
      }
    },
    "spec": {
      "selector": {
        "istio": "ingressgateway"
      },
      "servers": [
        {
          "port": {
            "number": 80,
            "name": "http",
            "protocol": "HTTP"
          },
          "hosts": [
            "spiffe-big-1.interussplatform.dev"
          ]
        }
      ]
    }
  },
  "istio-obj-45": {
    "apiVersion": "policy/v1beta1",
    "kind": "PodDisruptionBudget",
    "metadata": {
      "name": "ingressgateway",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-ingressgateway",
        "release": "istio",
        "istio": "ingressgateway"
      }
    },
    "spec": {
      "minAvailable": 1,
      "selector": {
        "matchLabels": {
          "app": "istio-ingressgateway",
          "release": "istio",
          "istio": "ingressgateway"
        }
      }
    }
  },
  "istio-obj-46": {
    "apiVersion": "rbac.authorization.k8s.io/v1",
    "kind": "Role",
    "metadata": {
      "name": "istio-ingressgateway-sds",
      "namespace": "istio-system",
      "labels": {
        "release": "istio"
      }
    },
    "rules": [
      {
        "apiGroups": [
          ""
        ],
        "resources": [
          "secrets"
        ],
        "verbs": [
          "get",
          "watch",
          "list"
        ]
      }
    ]
  },
  "istio-obj-47": {
    "apiVersion": "rbac.authorization.k8s.io/v1",
    "kind": "RoleBinding",
    "metadata": {
      "name": "istio-ingressgateway-sds",
      "namespace": "istio-system",
      "labels": {
        "release": "istio"
      }
    },
    "roleRef": {
      "apiGroup": "rbac.authorization.k8s.io",
      "kind": "Role",
      "name": "istio-ingressgateway-sds"
    },
    "subjects": [
      {
        "kind": "ServiceAccount",
        "name": "istio-ingressgateway-service-account"
      }
    ]
  },
  "istio-obj-48": {
    "apiVersion": "v1",
    "kind": "Service",
    "metadata": {
      "name": "istio-ingressgateway",
      "namespace": "istio-system",
      "annotations": null,
      "labels": {
        "app": "istio-ingressgateway",
        "release": "istio",
        "istio": "ingressgateway"
      }
    },
    "spec": {
      "type": "NodePort",
      "selector": {
        "app": "istio-ingressgateway"
      },
      "ports": [
        {
          "name": "status-port",
          "port": 15020,
          "targetPort": 15020
        },
        {
          "name": "http2",
          "port": 80,
          "targetPort": 80
        },
        {
          "name": "https",
          "port": 443
        },
        {
          "name": "kiali",
          "port": 15029,
          "targetPort": 15029
        },
        {
          "name": "prometheus",
          "port": 15030,
          "targetPort": 15030
        },
        {
          "name": "grafana",
          "port": 15031,
          "targetPort": 15031
        },
        {
          "name": "tracing",
          "port": 15032,
          "targetPort": 15032
        },
        {
          "name": "tls",
          "port": 15443,
          "targetPort": 15443
        }
      ]
    }
  },
  "istio-obj-49": {
    "apiVersion": "v1",
    "kind": "ServiceAccount",
    "metadata": {
      "name": "istio-ingressgateway-service-account",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-ingressgateway",
        "release": "istio"
      }
    }
  },
  "istio-obj-50": {
    "apiVersion": "networking.istio.io/v1alpha3",
    "kind": "Sidecar",
    "metadata": {
      "name": "default",
      "namespace": "istio-system",
      "labels": {
        "release": "istio"
      }
    },
    "spec": {
      "egress": [
        {
          "hosts": [
            "*/*"
          ]
        }
      ]
    }
  },
  "istio-obj-51": {
    "apiVersion": "rbac.authorization.k8s.io/v1",
    "kind": "ClusterRole",
    "metadata": {
      "name": "istio-sidecar-injector-istio-system",
      "labels": {
        "app": "sidecar-injector",
        "release": "istio",
        "istio": "sidecar-injector"
      }
    },
    "rules": [
      {
        "apiGroups": [
          ""
        ],
        "resources": [
          "configmaps"
        ],
        "resourceNames": [
          "istio-sidecar-injector"
        ],
        "verbs": [
          "get",
          "list",
          "watch"
        ]
      },
      {
        "apiGroups": [
          "admissionregistration.k8s.io"
        ],
        "resources": [
          "mutatingwebhookconfigurations"
        ],
        "resourceNames": [
          "istio-sidecar-injector",
          "istio-sidecar-injector-istio-system"
        ],
        "verbs": [
          "get",
          "list",
          "watch",
          "patch"
        ]
      }
    ]
  },
  "istio-obj-52": {
    "apiVersion": "rbac.authorization.k8s.io/v1",
    "kind": "ClusterRoleBinding",
    "metadata": {
      "name": "istio-sidecar-injector-admin-role-binding-istio-system",
      "labels": {
        "app": "sidecar-injector",
        "release": "istio",
        "istio": "sidecar-injector"
      }
    },
    "roleRef": {
      "apiGroup": "rbac.authorization.k8s.io",
      "kind": "ClusterRole",
      "name": "istio-sidecar-injector-istio-system"
    },
    "subjects": [
      {
        "kind": "ServiceAccount",
        "name": "istio-sidecar-injector-service-account",
        "namespace": "istio-system"
      }
    ]
  },
  "istio-obj-53": {
    "apiVersion": "v1",
    "kind": "ConfigMap",
    "metadata": {
      "name": "injector-mesh",
      "namespace": "istio-system",
      "labels": {
        "release": "istio"
      }
    },
    "data": {
      "mesh": "# Unix Domain Socket through which envoy communicates with NodeAgent SDS to get\n# key/cert for mTLS. Use secret-mount files instead of SDS if set to empty.\nsdsUdsPath: \"\"\n\ndefaultConfig:\n  #\n  # TCP connection timeout between Envoy & the application, and between Envoys.\n  connectTimeout: 10s\n  #\n  ### ADVANCED SETTINGS #############\n  # Where should envoy's configuration be stored in the istio-proxy container\n  configPath: \"/etc/istio/proxy\"\n  # The pseudo service name used for Envoy.\n  serviceCluster: istio-proxy\n  # These settings that determine how long an old Envoy\n  # process should be kept alive after an occasional reload.\n  drainDuration: 45s\n  parentShutdownDuration: 1m0s\n  #\n  # Port where Envoy listens (on local host) for admin commands\n  # You can exec into the istio-proxy container in a pod and\n  # curl the admin port (curl http://localhost:15000/) to obtain\n  # diagnostic information from Envoy. See\n  # https://lyft.github.io/envoy/docs/operations/admin.html\n  # for more details\n  proxyAdminPort: 15000\n  #\n  # Set concurrency to a specific number to control the number of Proxy worker threads.\n  # If set to 0 (default), then start worker thread for each CPU thread/core.\n  concurrency: 2\n  #\n  tracing:\n    zipkin:\n      # Address of the Zipkin collector\n      address: zipkin.istio-system:9411\n  #\n  # Mutual TLS authentication between sidecars and istio control plane.\n  controlPlaneAuthPolicy: MUTUAL_TLS\n  #\n  # Address where istio Pilot service is running\n  discoveryAddress: istio-pilot.istio-system:15011"
    }
  },
  "istio-obj-54": {
    "apiVersion": "apps/v1",
    "kind": "Deployment",
    "metadata": {
      "labels": {
        "app": "sidecarInjectorWebhook",
        "istio": "sidecar-injector",
        "release": "istio"
      },
      "name": "istio-sidecar-injector",
      "namespace": "istio-system"
    },
    "spec": {
      "replicas": 1,
      "selector": {
        "matchLabels": {
          "istio": "sidecar-injector"
        }
      },
      "strategy": {
        "rollingUpdate": {
          "maxSurge": "100%",
          "maxUnavailable": "25%"
        }
      },
      "template": {
        "metadata": {
          "annotations": {
            "sidecar.istio.io/inject": "false"
          },
          "labels": {
            "app": "sidecarInjectorWebhook",
            "chart": "sidecarInjectorWebhook",
            "heritage": "Tiller",
            "istio": "sidecar-injector",
            "release": "istio"
          }
        },
        "spec": {
          "affinity": {
            "nodeAffinity": {
              "preferredDuringSchedulingIgnoredDuringExecution": [
                {
                  "preference": {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "amd64"
                        ]
                      }
                    ]
                  },
                  "weight": 2
                },
                {
                  "preference": {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "ppc64le"
                        ]
                      }
                    ]
                  },
                  "weight": 2
                },
                {
                  "preference": {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "s390x"
                        ]
                      }
                    ]
                  },
                  "weight": 2
                }
              ],
              "requiredDuringSchedulingIgnoredDuringExecution": {
                "nodeSelectorTerms": [
                  {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "amd64",
                          "ppc64le",
                          "s390x"
                        ]
                      }
                    ]
                  }
                ]
              }
            }
          },
          "containers": [
            {
              "args": [
                "--caCertFile=/etc/istio/certs/root-cert.pem",
                "--tlsCertFile=/etc/istio/certs/cert-chain.pem",
                "--tlsKeyFile=/etc/istio/certs/key.pem",
                "--injectConfig=/etc/istio/inject/config",
                "--meshConfig=/etc/istio/config/mesh",
                "--port=9443",
                "--healthCheckInterval=2s",
                "--healthCheckFile=/tmp/health",
                "--reconcileWebhookConfig=true",
                "--webhookConfigName=istio-sidecar-injector",
                "--log_output_level=debug"
              ],
              "image": "docker.io/istio/sidecar_injector:1.4.3",
              "imagePullPolicy": "IfNotPresent",
              "livenessProbe": {
                "exec": {
                  "command": [
                    "/usr/local/bin/sidecar-injector",
                    "probe",
                    "--probe-path=/tmp/health",
                    "--interval=4s"
                  ]
                },
                "initialDelaySeconds": 4,
                "periodSeconds": 4
              },
              "name": "sidecar-injector-webhook",
              "readinessProbe": {
                "exec": {
                  "command": [
                    "/usr/local/bin/sidecar-injector",
                    "probe",
                    "--probe-path=/tmp/health",
                    "--interval=4s"
                  ]
                },
                "initialDelaySeconds": 4,
                "periodSeconds": 4
              },
              "resources": {
                "requests": {
                  "cpu": "10m"
                }
              },
              "volumeMounts": [
                {
                  "mountPath": "/etc/istio/config",
                  "name": "config-volume",
                  "readOnly": true
                },
                {
                  "mountPath": "/etc/istio/certs",
                  "name": "certs",
                  "readOnly": true
                },
                {
                  "mountPath": "/etc/istio/inject",
                  "name": "inject-config",
                  "readOnly": true
                }
              ]
            }
          ],
          "serviceAccountName": "istio-sidecar-injector-service-account",
          "volumes": [
            {
              "configMap": {
                "name": "injector-mesh"
              },
              "name": "config-volume"
            },
            {
              "name": "certs",
              "secret": {
                "secretName": "istio.istio-sidecar-injector-service-account"
              }
            },
            {
              "configMap": {
                "items": [
                  {
                    "key": "config",
                    "path": "config"
                  },
                  {
                    "key": "values",
                    "path": "values"
                  }
                ],
                "name": "istio-sidecar-injector"
              },
              "name": "inject-config"
            }
          ]
        }
      }
    }
  },
  "istio-obj-55": {
    "apiVersion": "admissionregistration.k8s.io/v1beta1",
    "kind": "MutatingWebhookConfiguration",
    "metadata": {
      "name": "istio-sidecar-injector",
      "labels": {
        "app": "sidecar-injector",
        "release": "istio"
      }
    },
    "webhooks": [
      {
        "name": "sidecar-injector.istio.io",
        "clientConfig": {
          "service": {
            "name": "istio-sidecar-injector",
            "namespace": "istio-system",
            "path": "/inject"
          },
          "caBundle": ""
        },
        "rules": [
          {
            "operations": [
              "CREATE"
            ],
            "apiGroups": [
              ""
            ],
            "apiVersions": [
              "v1"
            ],
            "resources": [
              "pods"
            ]
          }
        ],
        "failurePolicy": "Fail",
        "namespaceSelector": {
          "matchLabels": {
            "istio-injection": "enabled"
          }
        }
      }
    ]
  },
  "istio-obj-56": {
    "apiVersion": "policy/v1beta1",
    "kind": "PodDisruptionBudget",
    "metadata": {
      "name": "istio-sidecar-injector",
      "namespace": "istio-system",
      "labels": {
        "app": "sidecar-injector",
        "release": "istio",
        "istio": "sidecar-injector"
      }
    },
    "spec": {
      "minAvailable": 1,
      "selector": {
        "matchLabels": {
          "app": "sidecar-injector",
          "release": "istio",
          "istio": "sidecar-injector"
        }
      }
    }
  },
  "istio-obj-57": {
    "apiVersion": "v1",
    "kind": "Service",
    "metadata": {
      "name": "istio-sidecar-injector",
      "namespace": "istio-system",
      "labels": {
        "app": "sidecarInjectorWebhook",
        "release": "istio",
        "istio": "sidecar-injector"
      }
    },
    "spec": {
      "ports": [
        {
          "port": 443,
          "targetPort": 9443
        }
      ],
      "selector": {
        "istio": "sidecar-injector"
      }
    }
  },
  "istio-obj-58": {
    "apiVersion": "v1",
    "kind": "ServiceAccount",
    "metadata": {
      "name": "istio-sidecar-injector-service-account",
      "namespace": "istio-system",
      "labels": {
        "app": "sidecarInjectorWebhook",
        "release": "istio",
        "istio": "sidecar-injector"
      }
    }
  },
  "istio-obj-59": {
    "apiVersion": "v1",
    "kind": "ConfigMap",
    "metadata": {
      "name": "istio-sidecar-injector",
      "namespace": "istio-system",
      "labels": {
        "release": "istio",
        "app": "sidecar-injector",
        "istio": "sidecar-injector"
      }
    },
    "data": {
      "values": "{\"certmanager\":{\"enabled\":false,\"hub\":\"quay.io/jetstack\",\"image\":\"cert-manager-controller\",\"namespace\":\"istio-system\",\"tag\":\"v0.6.2\"},\"clusterResources\":true,\"cni\":{\"namespace\":\"istio-system\"},\"galley\":{\"enableAnalysis\":false,\"enabled\":true,\"image\":\"galley\",\"namespace\":\"istio-system\"},\"gateways\":{\"istio-egressgateway\":{\"autoscaleEnabled\":true,\"enabled\":false,\"env\":{\"ISTIO_META_ROUTER_MODE\":\"sni-dnat\"},\"namespace\":\"istio-system\",\"ports\":[{\"name\":\"http2\",\"port\":80},{\"name\":\"https\",\"port\":443},{\"name\":\"tls\",\"port\":15443,\"targetPort\":15443}],\"secretVolumes\":[{\"mountPath\":\"/etc/istio/egressgateway-certs\",\"name\":\"egressgateway-certs\",\"secretName\":\"istio-egressgateway-certs\"},{\"mountPath\":\"/etc/istio/egressgateway-ca-certs\",\"name\":\"egressgateway-ca-certs\",\"secretName\":\"istio-egressgateway-ca-certs\"}],\"type\":\"ClusterIP\",\"zvpn\":{\"enabled\":true,\"suffix\":\"global\"}},\"istio-ingressgateway\":{\"applicationPorts\":\"\",\"autoscaleEnabled\":true,\"debug\":\"info\",\"domain\":\"\",\"enabled\":true,\"env\":{\"ISTIO_META_ROUTER_MODE\":\"sni-dnat\"},\"meshExpansionPorts\":[{\"name\":\"tcp-pilot-grpc-tls\",\"port\":15011,\"targetPort\":15011},{\"name\":\"tcp-citadel-grpc-tls\",\"port\":8060,\"targetPort\":8060},{\"name\":\"tcp-dns-tls\",\"port\":853,\"targetPort\":853}],\"namespace\":\"istio-system\",\"ports\":[{\"name\":\"status-port\",\"port\":15020,\"targetPort\":15020},{\"name\":\"http2\",\"port\":80,\"targetPort\":80},{\"name\":\"https\",\"port\":443},{\"name\":\"kiali\",\"port\":15029,\"targetPort\":15029},{\"name\":\"prometheus\",\"port\":15030,\"targetPort\":15030},{\"name\":\"grafana\",\"port\":15031,\"targetPort\":15031},{\"name\":\"tracing\",\"port\":15032,\"targetPort\":15032},{\"name\":\"tls\",\"port\":15443,\"targetPort\":15443}],\"sds\":{\"enabled\":true,\"image\":\"node-agent-k8s\",\"resources\":{\"limits\":{\"cpu\":\"2000m\",\"memory\":\"1024Mi\"},\"requests\":{\"cpu\":\"100m\",\"memory\":\"128Mi\"}}},\"secretVolumes\":[{\"mountPath\":\"/etc/istio/ingressgateway-certs\",\"name\":\"ingressgateway-certs\",\"secretName\":\"istio-ingressgateway-certs\"},{\"mountPath\":\"/etc/istio/ingressgateway-ca-certs\",\"name\":\"ingressgateway-ca-certs\",\"secretName\":\"istio-ingressgateway-ca-certs\"}],\"type\":\"NodePort\",\"zvpn\":{\"enabled\":true,\"suffix\":\"global\"}}},\"global\":{\"arch\":{\"amd64\":2,\"ppc64le\":2,\"s390x\":2},\"certificates\":[],\"configNamespace\":\"istio-system\",\"configValidation\":true,\"controlPlaneSecurityEnabled\":true,\"defaultNodeSelector\":{},\"defaultPodDisruptionBudget\":{\"enabled\":true},\"defaultResources\":{\"requests\":{\"cpu\":\"10m\"}},\"disablePolicyChecks\":true,\"enableHelmTest\":false,\"enableTracing\":true,\"enabled\":true,\"hub\":\"docker.io/istio\",\"imagePullPolicy\":\"IfNotPresent\",\"imagePullSecrets\":[],\"istioNamespace\":\"istio-system\",\"k8sIngress\":{\"enableHttps\":false,\"enabled\":false,\"gatewayName\":\"ingressgateway\"},\"localityLbSetting\":{\"enabled\":true},\"logAsJson\":false,\"logging\":{\"level\":\"default:info\"},\"meshExpansion\":{\"enabled\":false,\"useILB\":false},\"meshNetworks\":{},\"mtls\":{\"auto\":false,\"enabled\":false},\"multiCluster\":{\"clusterName\":\"\",\"enabled\":false},\"namespace\":\"istio-system\",\"network\":\"\",\"omitSidecarInjectorConfigMap\":false,\"oneNamespace\":false,\"operatorManageWebhooks\":false,\"outboundTrafficPolicy\":{\"mode\":\"ALLOW_ANY\"},\"policyCheckFailOpen\":false,\"policyNamespace\":\"istio-system\",\"priorityClassName\":\"\",\"prometheusNamespace\":\"istio-system\",\"proxy\":{\"accessLogEncoding\":\"TEXT\",\"accessLogFile\":\"\",\"accessLogFormat\":\"\",\"autoInject\":\"enabled\",\"clusterDomain\":\"cluster.local\",\"componentLogLevel\":\"misc:error\",\"concurrency\":2,\"dnsRefreshRate\":\"300s\",\"enableCoreDump\":false,\"envoyAccessLogService\":{\"enabled\":false},\"envoyMetricsService\":{\"enabled\":false,\"tcpKeepalive\":{\"interval\":\"10s\",\"probes\":3,\"time\":\"10s\"},\"tlsSettings\":{\"mode\":\"DISABLE\",\"subjectAltNames\":[]}},\"envoyStatsd\":{\"enabled\":false},\"excludeIPRanges\":\"\",\"excludeInboundPorts\":\"\",\"excludeOutboundPorts\":\"\",\"image\":\"proxyv2\",\"includeIPRanges\":\"*\",\"includeInboundPorts\":\"*\",\"kubevirtInterfaces\":\"\",\"logLevel\":\"warning\",\"privileged\":false,\"protocolDetectionTimeout\":\"100ms\",\"readinessFailureThreshold\":30,\"readinessInitialDelaySeconds\":1,\"readinessPeriodSeconds\":2,\"resources\":{\"limits\":{\"cpu\":\"2000m\",\"memory\":\"1024Mi\"},\"requests\":{\"cpu\":\"100m\",\"memory\":\"128Mi\"}},\"statusPort\":15020,\"tracer\":\"zipkin\"},\"proxy_init\":{\"image\":\"proxyv2\",\"resources\":{\"limits\":{\"cpu\":\"100m\",\"memory\":\"50Mi\"},\"requests\":{\"cpu\":\"10m\",\"memory\":\"10Mi\"}}},\"sds\":{\"enabled\":false,\"token\":{\"aud\":\"istio-ca\"},\"udsPath\":\"\"},\"securityNamespace\":\"istio-system\",\"tag\":\"1.4.3\",\"telemetryNamespace\":\"istio-system\",\"tracer\":{\"datadog\":{\"address\":\"$(HOST_IP):8126\"},\"lightstep\":{\"accessToken\":\"\",\"address\":\"\",\"cacertPath\":\"\",\"secure\":true},\"zipkin\":{\"address\":\"\"}},\"trustDomain\":\"cluster.local\",\"useMCP\":true},\"grafana\":{\"accessMode\":\"ReadWriteMany\",\"contextPath\":\"/grafana\",\"dashboardProviders\":{\"dashboardproviders.yaml\":{\"apiVersion\":1,\"providers\":[{\"disableDeletion\":false,\"folder\":\"istio\",\"name\":\"istio\",\"options\":{\"path\":\"/var/lib/grafana/dashboards/istio\"},\"orgId\":1,\"type\":\"file\"}]}},\"datasources\":{\"datasources.yaml\":{\"apiVersion\":1}},\"enabled\":false,\"env\":{},\"envSecrets\":{},\"image\":{\"repository\":\"grafana/grafana\",\"tag\":\"6.4.3\"},\"ingress\":{\"enabled\":false,\"hosts\":[\"grafana.local\"]},\"namespace\":\"istio-system\",\"nodeSelector\":{},\"persist\":false,\"podAntiAffinityLabelSelector\":[],\"podAntiAffinityTermLabelSelector\":[],\"replicaCount\":1,\"security\":{\"enabled\":false,\"passphraseKey\":\"passphrase\",\"secretName\":\"grafana\",\"usernameKey\":\"username\"},\"service\":{\"annotations\":{},\"externalPort\":3000,\"name\":\"http\",\"type\":\"ClusterIP\"},\"storageClassName\":\"\",\"tolerations\":[]},\"istio_cni\":{\"enabled\":false},\"istiocoredns\":{\"coreDNSImage\":\"coredns/coredns\",\"coreDNSPluginImage\":\"istio/coredns-plugin:0.2-istio-1.1\",\"coreDNSTag\":\"1.6.2\",\"enabled\":false,\"namespace\":\"istio-system\"},\"kiali\":{\"contextPath\":\"/kiali\",\"createDemoSecret\":false,\"dashboard\":{\"passphraseKey\":\"passphrase\",\"secretName\":\"kiali\",\"usernameKey\":\"username\",\"viewOnlyMode\":false},\"enabled\":false,\"hub\":\"quay.io/kiali\",\"ingress\":{\"enabled\":false,\"hosts\":[\"kiali.local\"]},\"namespace\":\"istio-system\",\"nodeSelector\":{},\"podAntiAffinityLabelSelector\":[],\"podAntiAffinityTermLabelSelector\":[],\"replicaCount\":1,\"security\":{\"cert_file\":\"/kiali-cert/cert-chain.pem\",\"enabled\":false,\"private_key_file\":\"/kiali-cert/key.pem\"},\"tag\":\"v1.9\"},\"mixer\":{\"adapters\":{\"kubernetesenv\":{\"enabled\":true},\"prometheus\":{\"enabled\":true,\"metricsExpiryDuration\":\"10m\"},\"stackdriver\":{\"auth\":{\"apiKey\":\"\",\"appCredentials\":false,\"serviceAccountPath\":\"\"},\"enabled\":false,\"tracer\":{\"enabled\":false,\"sampleProbability\":1}},\"stdio\":{\"enabled\":false,\"outputAsJson\":false},\"useAdapterCRDs\":false},\"policy\":{\"adapters\":{\"kubernetesenv\":{\"enabled\":true},\"useAdapterCRDs\":false},\"autoscaleEnabled\":true,\"enabled\":true,\"image\":\"mixer\",\"namespace\":\"istio-system\",\"sessionAffinityEnabled\":false},\"telemetry\":{\"autoscaleEnabled\":true,\"enabled\":true,\"env\":{\"GOMAXPROCS\":\"6\"},\"image\":\"mixer\",\"loadshedding\":{\"latencyThreshold\":\"100ms\",\"mode\":\"enforce\"},\"namespace\":\"istio-system\",\"nodeSelector\":{},\"podAntiAffinityLabelSelector\":[],\"podAntiAffinityTermLabelSelector\":[],\"replicaCount\":1,\"reportBatchMaxEntries\":100,\"reportBatchMaxTime\":\"1s\",\"sessionAffinityEnabled\":false,\"tolerations\":[],\"useMCP\":true}},\"nodeagent\":{\"enabled\":false,\"image\":\"node-agent-k8s\",\"namespace\":\"istio-system\"},\"pilot\":{\"appNamespaces\":[],\"autoscaleEnabled\":true,\"autoscaleMax\":5,\"autoscaleMin\":1,\"configMap\":true,\"configNamespace\":\"istio-config\",\"cpu\":{\"targetAverageUtilization\":80},\"enableProtocolSniffingForInbound\":false,\"enableProtocolSniffingForOutbound\":true,\"enabled\":true,\"env\":{},\"image\":\"pilot\",\"ingress\":{\"ingressClass\":\"istio\",\"ingressControllerMode\":\"OFF\",\"ingressService\":\"istio-ingressgateway\"},\"keepaliveMaxServerConnectionAge\":\"30m\",\"meshNetworks\":{\"networks\":{}},\"namespace\":\"istio-system\",\"nodeSelector\":{},\"podAntiAffinityLabelSelector\":[],\"podAntiAffinityTermLabelSelector\":[],\"policy\":{\"enabled\":false},\"replicaCount\":1,\"tolerations\":[],\"traceSampling\":1,\"useMCP\":true},\"prometheus\":{\"contextPath\":\"/prometheus\",\"enabled\":true,\"hub\":\"docker.io/prom\",\"ingress\":{\"enabled\":false,\"hosts\":[\"prometheus.local\"]},\"namespace\":\"istio-system\",\"nodeSelector\":{},\"podAntiAffinityLabelSelector\":[],\"podAntiAffinityTermLabelSelector\":[],\"replicaCount\":1,\"retention\":\"6h\",\"scrapeInterval\":\"15s\",\"security\":{\"enabled\":true},\"tag\":\"v2.12.0\",\"tolerations\":[]},\"security\":{\"dnsCerts\":{\"istio-pilot-service-account.istio-control\":\"istio-pilot.istio-control\"},\"enableNamespacesByDefault\":true,\"enabled\":true,\"image\":\"citadel\",\"namespace\":\"istio-system\",\"selfSigned\":true,\"trustDomain\":\"cluster.local\"},\"sidecarInjectorWebhook\":{\"alwaysInjectSelector\":[],\"enableNamespacesByDefault\":false,\"enabled\":true,\"image\":\"sidecar_injector\",\"injectLabel\":\"istio-injection\",\"injectedAnnotations\":{},\"namespace\":\"istio-system\",\"neverInjectSelector\":[],\"nodeSelector\":{},\"objectSelector\":{\"autoInject\":true,\"enabled\":false},\"podAnnotations\":{},\"podAntiAffinityLabelSelector\":[],\"podAntiAffinityTermLabelSelector\":[],\"replicaCount\":1,\"resources\":{},\"rewriteAppHTTPProbe\":false,\"rollingMaxSurge\":\"100%\",\"rollingMaxUnavailable\":\"25%\",\"selfSigned\":false,\"tolerations\":[]},\"telemetry\":{\"enabled\":true,\"v1\":{\"enabled\":true},\"v2\":{\"enabled\":false,\"prometheus\":{\"enabled\":true},\"stackdriver\":{\"configOverride\":{},\"enabled\":false,\"logging\":false,\"monitoring\":false,\"topology\":false}}},\"tracing\":{\"enabled\":false,\"ingress\":{\"enabled\":false},\"jaeger\":{\"accessMode\":\"ReadWriteMany\",\"enabled\":false,\"hub\":\"docker.io/jaegertracing\",\"memory\":{\"max_traces\":50000},\"namespace\":\"istio-system\",\"persist\":false,\"spanStorageType\":\"badger\",\"storageClassName\":\"\",\"tag\":\"1.14\"},\"nodeSelector\":{},\"opencensus\":{\"exporters\":{\"stackdriver\":{\"enable_tracing\":true}},\"hub\":\"docker.io/omnition\",\"resources\":{\"limits\":{\"cpu\":\"1\",\"memory\":\"2Gi\"},\"requests\":{\"cpu\":\"200m\",\"memory\":\"400Mi\"}},\"tag\":\"0.1.9\"},\"podAntiAffinityLabelSelector\":[],\"podAntiAffinityTermLabelSelector\":[],\"provider\":\"jaeger\",\"service\":{\"annotations\":{},\"externalPort\":9411,\"name\":\"http-query\",\"type\":\"ClusterIP\"},\"zipkin\":{\"hub\":\"docker.io/openzipkin\",\"javaOptsHeap\":700,\"maxSpans\":500000,\"node\":{\"cpus\":2},\"probeStartupDelay\":200,\"queryPort\":9411,\"resources\":{\"limits\":{\"cpu\":\"300m\",\"memory\":\"900Mi\"},\"requests\":{\"cpu\":\"150m\",\"memory\":\"900Mi\"}},\"tag\":\"2.14.2\"}},\"version\":\"\"}",
      "config": "policy: enabled\nalwaysInjectSelector:\n  []\nneverInjectSelector:\n  []\ntemplate: |\n  rewriteAppHTTPProbe: {{ valueOrDefault .Values.sidecarInjectorWebhook.rewriteAppHTTPProbe false }}\n  {{- if or (not .Values.istio_cni.enabled) .Values.global.proxy.enableCoreDump }}\n  initContainers:\n  {{ if ne (annotation .ObjectMeta `sidecar.istio.io/interceptionMode` .ProxyConfig.InterceptionMode) `NONE` }}\n  {{- if not .Values.istio_cni.enabled }}\n  - name: istio-init\n  {{- if contains \"/\" .Values.global.proxy_init.image }}\n    image: \"{{ .Values.global.proxy_init.image }}\"\n  {{- else }}\n    image: \"{{ .Values.global.hub }}/{{ .Values.global.proxy_init.image }}:{{ .Values.global.tag }}\"\n  {{- end }}\n    command:\n    - istio-iptables\n    - \"-p\"\n    - 15001\n    - \"-z\"\n    - \"15006\"\n    - \"-u\"\n    - 1337\n    - \"-m\"\n    - \"{{ annotation .ObjectMeta `sidecar.istio.io/interceptionMode` .ProxyConfig.InterceptionMode }}\"\n    - \"-i\"\n    - \"{{ annotation .ObjectMeta `traffic.sidecar.istio.io/includeOutboundIPRanges` .Values.global.proxy.includeIPRanges }}\"\n    - \"-x\"\n    - \"{{ annotation .ObjectMeta `traffic.sidecar.istio.io/excludeOutboundIPRanges` .Values.global.proxy.excludeIPRanges }}\"\n    - \"-b\"\n    - \"{{ annotation .ObjectMeta `traffic.sidecar.istio.io/includeInboundPorts` `*` }}\"\n    - \"-d\"\n    - \"{{ excludeInboundPort (annotation .ObjectMeta `status.sidecar.istio.io/port` .Values.global.proxy.statusPort) (annotation .ObjectMeta `traffic.sidecar.istio.io/excludeInboundPorts` .Values.global.proxy.excludeInboundPorts) }}\"\n    {{ if or (isset .ObjectMeta.Annotations `traffic.sidecar.istio.io/excludeOutboundPorts`) (ne (valueOrDefault .Values.global.proxy.excludeOutboundPorts \"\") \"\") -}}\n    - \"-o\"\n    - \"{{ annotation .ObjectMeta `traffic.sidecar.istio.io/excludeOutboundPorts` .Values.global.proxy.excludeOutboundPorts }}\"\n    {{ end -}}\n    {{ if (isset .ObjectMeta.Annotations `traffic.sidecar.istio.io/kubevirtInterfaces`) -}}\n    - \"-k\"\n    - \"{{ index .ObjectMeta.Annotations `traffic.sidecar.istio.io/kubevirtInterfaces` }}\"\n    {{ end -}}\n    imagePullPolicy: \"{{ valueOrDefault .Values.global.imagePullPolicy `Always` }}\"\n  {{- if .Values.global.proxy_init.resources }}\n    resources:\n      {{ toYaml .Values.global.proxy_init.resources | indent 4 }}\n  {{- else }}\n    resources: {}\n  {{- end }}\n    securityContext:\n      runAsUser: 0\n      runAsNonRoot: false\n      capabilities:\n        add:\n        - NET_ADMIN\n      {{- if .Values.global.proxy.privileged }}\n      privileged: true\n      {{- end }}\n    restartPolicy: Always\n  {{- end }}\n  {{  end -}}\n  {{- if eq .Values.global.proxy.enableCoreDump true }}\n  - name: enable-core-dump\n    args:\n    - -c\n    - sysctl -w kernel.core_pattern=/var/lib/istio/core.proxy && ulimit -c unlimited\n    command:\n      - /bin/sh\n  {{- if contains \"/\" .Values.global.proxy_init.image }}\n    image: \"{{ .Values.global.proxy_init.image }}\"\n  {{- else }}\n    image: \"{{ .Values.global.hub }}/{{ .Values.global.proxy_init.image }}:{{ .Values.global.tag }}\"\n  {{- end }}\n    imagePullPolicy: \"{{ valueOrDefault .Values.global.imagePullPolicy `Always` }}\"\n    resources: {}\n    securityContext:\n      runAsUser: 0\n      runAsNonRoot: false\n      privileged: true\n  {{ end }}\n  {{- end }}\n  containers:\n  - name: istio-proxy\n  {{- if contains \"/\" (annotation .ObjectMeta `sidecar.istio.io/proxyImage` .Values.global.proxy.image) }}\n    image: \"{{ annotation .ObjectMeta `sidecar.istio.io/proxyImage` .Values.global.proxy.image }}\"\n  {{- else }}\n    image: \"{{ .Values.global.hub }}/{{ .Values.global.proxy.image }}:{{ .Values.global.tag }}\"\n  {{- end }}\n    ports:\n    - containerPort: 15090\n      protocol: TCP\n      name: http-envoy-prom\n    args:\n    - proxy\n    - sidecar\n    - --domain\n    - $(POD_NAMESPACE).svc.{{ .Values.global.proxy.clusterDomain }}\n    - --configPath\n    - \"/etc/istio/proxy\"\n    - --binaryPath\n    - \"/usr/local/bin/envoy\"\n    - --serviceCluster\n    {{ if ne \"\" (index .ObjectMeta.Labels \"app\") -}}\n    - \"{{ index .ObjectMeta.Labels `app` }}.$(POD_NAMESPACE)\"\n    {{ else -}}\n    - \"{{ valueOrDefault .DeploymentMeta.Name `istio-proxy` }}.{{ valueOrDefault .DeploymentMeta.Namespace `default` }}\"\n    {{ end -}}\n    - --drainDuration\n    - \"{{ formatDuration .ProxyConfig.DrainDuration }}\"\n    - --parentShutdownDuration\n    - \"{{ formatDuration .ProxyConfig.ParentShutdownDuration }}\"\n    - --discoveryAddress\n    - \"{{ annotation .ObjectMeta `sidecar.istio.io/discoveryAddress` .ProxyConfig.DiscoveryAddress }}\"\n  {{- if eq .Values.global.proxy.tracer \"lightstep\" }}\n    - --lightstepAddress\n    - \"{{ .ProxyConfig.GetTracing.GetLightstep.GetAddress }}\"\n    - --lightstepAccessToken\n    - \"{{ .ProxyConfig.GetTracing.GetLightstep.GetAccessToken }}\"\n    - --lightstepSecure={{ .ProxyConfig.GetTracing.GetLightstep.GetSecure }}\n    - --lightstepCacertPath\n    - \"{{ .ProxyConfig.GetTracing.GetLightstep.GetCacertPath }}\"\n  {{- else if eq .Values.global.proxy.tracer \"zipkin\" }}\n    - --zipkinAddress\n    - \"{{ .ProxyConfig.GetTracing.GetZipkin.GetAddress }}\"\n  {{- else if eq .Values.global.proxy.tracer \"datadog\" }}\n    - --datadogAgentAddress\n    - \"{{ .ProxyConfig.GetTracing.GetDatadog.GetAddress }}\"\n  {{- end }}\n    - --proxyLogLevel={{ annotation .ObjectMeta `sidecar.istio.io/logLevel` .Values.global.proxy.logLevel}}\n    - --proxyComponentLogLevel={{ annotation .ObjectMeta `sidecar.istio.io/componentLogLevel` .Values.global.proxy.componentLogLevel}}\n    - --connectTimeout\n    - \"{{ formatDuration .ProxyConfig.ConnectTimeout }}\"\n  {{- if .Values.global.proxy.envoyStatsd.enabled }}\n    - --statsdUdpAddress\n    - \"{{ .ProxyConfig.StatsdUdpAddress }}\"\n  {{- end }}\n  {{- if .Values.global.proxy.envoyMetricsService.enabled }}\n    - --envoyMetricsServiceAddress\n    - \"{{ .ProxyConfig.GetEnvoyMetricsService.GetAddress }}\"\n  {{- end }}\n  {{- if .Values.global.proxy.envoyAccessLogService.enabled }}\n    - --envoyAccessLogServiceAddress\n    - \"{{ .ProxyConfig.GetEnvoyAccessLogService.GetAddress }}\"\n  {{- end }}\n    - --proxyAdminPort\n    - \"{{ .ProxyConfig.ProxyAdminPort }}\"\n    {{ if gt .ProxyConfig.Concurrency 0 -}}\n    - --concurrency\n    - \"{{ .ProxyConfig.Concurrency }}\"\n    {{ end -}}\n    {{- if .Values.global.controlPlaneSecurityEnabled }}\n    - --controlPlaneAuthPolicy\n    - MUTUAL_TLS\n    {{- else }}\n    - --controlPlaneAuthPolicy\n    - NONE\n    {{- end }}\n    - --dnsRefreshRate\n    - {{ valueOrDefault .Values.global.proxy.dnsRefreshRate \"300s\" }}\n  {{- if (ne (annotation .ObjectMeta \"status.sidecar.istio.io/port\" .Values.global.proxy.statusPort) \"0\") }}\n    - --statusPort\n    - \"{{ annotation .ObjectMeta `status.sidecar.istio.io/port` .Values.global.proxy.statusPort }}\"\n    - --applicationPorts\n    - \"{{ annotation .ObjectMeta `readiness.status.sidecar.istio.io/applicationPorts` (applicationPorts .Spec.Containers) }}\"\n  {{- end }}\n  {{- if .Values.global.trustDomain }}\n    - --trust-domain={{ .Values.global.trustDomain }}\n  {{- end }}\n  {{- if .Values.global.logAsJson }}\n    - --log_as_json\n  {{- end }}\n  {{- if (isset .ObjectMeta.Annotations `sidecar.istio.io/bootstrapOverride`) }}\n    - --templateFile=/etc/istio/custom-bootstrap/envoy_bootstrap.json\n  {{- end }}\n    env:\n    - name: POD_NAME\n      valueFrom:\n        fieldRef:\n          fieldPath: metadata.name\n    - name: POD_NAMESPACE\n      valueFrom:\n        fieldRef:\n          fieldPath: metadata.namespace\n    - name: INSTANCE_IP\n      valueFrom:\n        fieldRef:\n          fieldPath: status.podIP\n    - name: SERVICE_ACCOUNT\n      valueFrom:\n        fieldRef:\n          fieldPath: spec.serviceAccountName\n    - name: HOST_IP\n      valueFrom:\n        fieldRef:\n          fieldPath: status.hostIP\n  {{- if eq .Values.global.proxy.tracer \"datadog\" }}\n  {{- if isset .ObjectMeta.Annotations `apm.datadoghq.com/env` }}\n  {{- range $key, $value := fromJSON (index .ObjectMeta.Annotations `apm.datadoghq.com/env`) }}\n    - name: {{ $key }}\n      value: \"{{ $value }}\"\n  {{- end }}\n  {{- end }}\n  {{- end }}\n    - name: ISTIO_META_POD_PORTS\n      value: |-\n        [\n        {{- $first := true }}\n        {{- range $index1, $c := .Spec.Containers }}\n          {{- range $index2, $p := $c.Ports }}\n            {{- if (structToJSON $p) }}\n            {{if not $first}},{{end}}{{ structToJSON $p }}\n            {{- $first = false }}\n            {{- end }}\n          {{- end}}\n        {{- end}}\n        ]\n    - name: ISTIO_META_CLUSTER_ID\n      value: \"{{ valueOrDefault .Values.global.multiCluster.clusterName `Kubernetes` }}\"\n    - name: ISTIO_META_POD_NAME\n      valueFrom:\n        fieldRef:\n          fieldPath: metadata.name\n    - name: ISTIO_META_CONFIG_NAMESPACE\n      valueFrom:\n        fieldRef:\n          fieldPath: metadata.namespace\n    - name: SDS_ENABLED\n      value: \"{{ .Values.global.sds.enabled }}\"\n    - name: ISTIO_META_INTERCEPTION_MODE\n      value: \"{{ or (index .ObjectMeta.Annotations `sidecar.istio.io/interceptionMode`) .ProxyConfig.InterceptionMode.String }}\"\n    - name: ISTIO_META_INCLUDE_INBOUND_PORTS\n      value: \"{{ annotation .ObjectMeta `traffic.sidecar.istio.io/includeInboundPorts` (applicationPorts .Spec.Containers) }}\"\n    {{- if .Values.global.network }}\n    - name: ISTIO_META_NETWORK\n      value: \"{{ .Values.global.network }}\"\n    {{- end }}\n    {{ if .ObjectMeta.Annotations }}\n    - name: ISTIO_METAJSON_ANNOTATIONS\n      value: |\n             {{ toJSON .ObjectMeta.Annotations }}\n    {{ end }}\n    {{ if .ObjectMeta.Labels }}\n    - name: ISTIO_METAJSON_LABELS\n      value: |\n             {{ toJSON .ObjectMeta.Labels }}\n    {{ end }}\n    {{- if .DeploymentMeta.Name }}\n    - name: ISTIO_META_WORKLOAD_NAME\n      value: {{ .DeploymentMeta.Name }}\n    {{ end }}\n    {{- if and .TypeMeta.APIVersion .DeploymentMeta.Name }}\n    - name: ISTIO_META_OWNER\n      value: kubernetes://apis/{{ .TypeMeta.APIVersion }}/namespaces/{{ valueOrDefault .DeploymentMeta.Namespace `default` }}/{{ toLower .TypeMeta.Kind}}s/{{ .DeploymentMeta.Name }}\n    {{- end}}\n    {{- if (isset .ObjectMeta.Annotations `sidecar.istio.io/bootstrapOverride`) }}\n    - name: ISTIO_BOOTSTRAP_OVERRIDE\n      value: \"/etc/istio/custom-bootstrap/custom_bootstrap.json\"\n    {{- end }}\n    {{- if .Values.global.sds.customTokenDirectory }}\n    - name: ISTIO_META_SDS_TOKEN_PATH\n      value: \"{{ .Values.global.sds.customTokenDirectory -}}/sdstoken\"\n    {{- end }}\n    {{- if .Values.global.meshID }}\n    - name: ISTIO_META_MESH_ID\n      value: \"{{ .Values.global.meshID }}\"\n    {{- else if .Values.global.trustDomain }}\n    - name: ISTIO_META_MESH_ID\n      value: \"{{ .Values.global.trustDomain }}\"\n    {{- end }}\n    {{- if and (eq .Values.global.proxy.tracer \"datadog\") (isset .ObjectMeta.Annotations `apm.datadoghq.com/env`) }}\n    {{- range $key, $value := fromJSON (index .ObjectMeta.Annotations `apm.datadoghq.com/env`) }}\n      - name: {{ $key }}\n        value: \"{{ $value }}\"\n    {{- end }}\n    {{- end }}\n    imagePullPolicy: \"{{ valueOrDefault .Values.global.imagePullPolicy `Always` }}\"\n    {{ if ne (annotation .ObjectMeta `status.sidecar.istio.io/port` .Values.global.proxy.statusPort) `0` }}\n    readinessProbe:\n      httpGet:\n        path: /healthz/ready\n        port: {{ annotation .ObjectMeta `status.sidecar.istio.io/port` .Values.global.proxy.statusPort }}\n      initialDelaySeconds: {{ annotation .ObjectMeta `readiness.status.sidecar.istio.io/initialDelaySeconds` .Values.global.proxy.readinessInitialDelaySeconds }}\n      periodSeconds: {{ annotation .ObjectMeta `readiness.status.sidecar.istio.io/periodSeconds` .Values.global.proxy.readinessPeriodSeconds }}\n      failureThreshold: {{ annotation .ObjectMeta `readiness.status.sidecar.istio.io/failureThreshold` .Values.global.proxy.readinessFailureThreshold }}\n    {{ end -}}\n    securityContext:\n      {{- if .Values.global.proxy.privileged }}\n      privileged: true\n      {{- end }}\n      {{- if ne .Values.global.proxy.enableCoreDump true }}\n      readOnlyRootFilesystem: true\n      {{- end }}\n      {{ if eq (annotation .ObjectMeta `sidecar.istio.io/interceptionMode` .ProxyConfig.InterceptionMode) `TPROXY` -}}\n      capabilities:\n        add:\n        - NET_ADMIN\n      runAsGroup: 1337\n      {{ else -}}\n      {{ if .Values.global.sds.enabled }}\n      runAsGroup: 1337\n      {{- end }}\n      runAsUser: 1337\n      {{- end }}\n    resources:\n      {{ if or (isset .ObjectMeta.Annotations `sidecar.istio.io/proxyCPU`) (isset .ObjectMeta.Annotations `sidecar.istio.io/proxyMemory`) -}}\n      requests:\n        {{ if (isset .ObjectMeta.Annotations `sidecar.istio.io/proxyCPU`) -}}\n        cpu: \"{{ index .ObjectMeta.Annotations `sidecar.istio.io/proxyCPU` }}\"\n        {{ end}}\n        {{ if (isset .ObjectMeta.Annotations `sidecar.istio.io/proxyMemory`) -}}\n        memory: \"{{ index .ObjectMeta.Annotations `sidecar.istio.io/proxyMemory` }}\"\n        {{ end }}\n    {{ else -}}\n  {{- if .Values.global.proxy.resources }}\n      {{ toYaml .Values.global.proxy.resources | indent 4 }}\n  {{- end }}\n    {{  end -}}\n    volumeMounts:\n    {{ if (isset .ObjectMeta.Annotations `sidecar.istio.io/bootstrapOverride`) }}\n    - mountPath: /etc/istio/custom-bootstrap\n      name: custom-bootstrap-volume\n    {{- end }}\n    - mountPath: /etc/istio/proxy\n      name: istio-envoy\n    {{- if .Values.global.sds.enabled }}\n    - mountPath: /var/run/sds\n      name: sds-uds-path\n      readOnly: true\n    - mountPath: /var/run/secrets/tokens\n      name: istio-token\n    {{- if .Values.global.sds.customTokenDirectory }}\n    - mountPath: \"{{ .Values.global.sds.customTokenDirectory -}}\"\n      name: custom-sds-token\n      readOnly: true\n    {{- end }}\n    {{- else }}\n    - mountPath: /etc/certs/\n      name: istio-certs\n      readOnly: true\n    {{- end }}\n    {{- if and (eq .Values.global.proxy.tracer \"lightstep\") .Values.global.tracer.lightstep.cacertPath }}\n    - mountPath: {{ directory .ProxyConfig.GetTracing.GetLightstep.GetCacertPath }}\n      name: lightstep-certs\n      readOnly: true\n    {{- end }}\n      {{- if isset .ObjectMeta.Annotations `sidecar.istio.io/userVolumeMount` }}\n      {{ range $index, $value := fromJSON (index .ObjectMeta.Annotations `sidecar.istio.io/userVolumeMount`) }}\n    - name: \"{{  $index }}\"\n      {{ toYaml $value | indent 4 }}\n      {{ end }}\n      {{- end }}\n  volumes:\n  {{- if (isset .ObjectMeta.Annotations `sidecar.istio.io/bootstrapOverride`) }}\n  - name: custom-bootstrap-volume\n    configMap:\n      name: {{ annotation .ObjectMeta `sidecar.istio.io/bootstrapOverride` \"\" }}\n  {{- end }}\n  - emptyDir:\n      medium: Memory\n    name: istio-envoy\n  {{- if .Values.global.sds.enabled }}\n  - name: sds-uds-path\n    hostPath:\n      path: /var/run/sds\n  - name: istio-token\n    projected:\n      sources:\n      - serviceAccountToken:\n          path: istio-token\n          expirationSeconds: 43200\n          audience: {{ .Values.global.sds.token.aud }}\n  {{- if .Values.global.sds.customTokenDirectory }}\n  - name: custom-sds-token\n    secret:\n      secretName: sdstokensecret\n  {{- end }}\n  {{- else }}\n  - name: istio-certs\n    secret:\n      optional: true\n      {{ if eq .Spec.ServiceAccountName \"\" }}\n      secretName: istio.default\n      {{ else -}}\n      secretName: {{  printf \"istio.%s\" .Spec.ServiceAccountName }}\n      {{  end -}}\n    {{- if isset .ObjectMeta.Annotations `sidecar.istio.io/userVolume` }}\n    {{range $index, $value := fromJSON (index .ObjectMeta.Annotations `sidecar.istio.io/userVolume`) }}\n  - name: \"{{ $index }}\"\n    {{ toYaml $value | indent 2 }}\n    {{ end }}\n    {{ end }}\n  {{- end }}\n  {{- if and (eq .Values.global.proxy.tracer \"lightstep\") .Values.global.tracer.lightstep.cacertPath }}\n  - name: lightstep-certs\n    secret:\n      optional: true\n      secretName: lightstep.cacert\n  {{- end }}\n  {{- if .Values.global.podDNSSearchNamespaces }}\n  dnsConfig:\n    searches:\n      {{- range .Values.global.podDNSSearchNamespaces }}\n      - {{ render . }}\n      {{- end }}\n  {{- end }}\ninjectedAnnotations:"
    }
  },
  "istio-obj-60": {
    "apiVersion": "autoscaling/v2beta1",
    "kind": "HorizontalPodAutoscaler",
    "metadata": {
      "labels": {
        "app": "pilot",
        "release": "istio"
      },
      "name": "istio-pilot",
      "namespace": "istio-system"
    },
    "spec": {
      "maxReplicas": 5,
      "metrics": [
        {
          "resource": {
            "name": "cpu",
            "targetAverageUtilization": 80
          },
          "type": "Resource"
        }
      ],
      "minReplicas": 1,
      "scaleTargetRef": {
        "apiVersion": "apps/v1",
        "kind": "Deployment",
        "name": "istio-pilot"
      }
    }
  },
  "istio-obj-61": {
    "apiVersion": "rbac.authorization.k8s.io/v1",
    "kind": "ClusterRole",
    "metadata": {
      "name": "istio-pilot-istio-system",
      "labels": {
        "app": "pilot",
        "release": "istio"
      }
    },
    "rules": [
      {
        "apiGroups": [
          "config.istio.io"
        ],
        "resources": [
          "*"
        ],
        "verbs": [
          "*"
        ]
      },
      {
        "apiGroups": [
          "rbac.istio.io"
        ],
        "resources": [
          "*"
        ],
        "verbs": [
          "get",
          "watch",
          "list"
        ]
      },
      {
        "apiGroups": [
          "security.istio.io"
        ],
        "resources": [
          "*"
        ],
        "verbs": [
          "get",
          "watch",
          "list"
        ]
      },
      {
        "apiGroups": [
          "networking.istio.io"
        ],
        "resources": [
          "*"
        ],
        "verbs": [
          "*"
        ]
      },
      {
        "apiGroups": [
          "authentication.istio.io"
        ],
        "resources": [
          "*"
        ],
        "verbs": [
          "*"
        ]
      },
      {
        "apiGroups": [
          "apiextensions.k8s.io"
        ],
        "resources": [
          "customresourcedefinitions"
        ],
        "verbs": [
          "*"
        ]
      },
      {
        "apiGroups": [
          "extensions"
        ],
        "resources": [
          "ingresses",
          "ingresses/status"
        ],
        "verbs": [
          "*"
        ]
      },
      {
        "apiGroups": [
          ""
        ],
        "resources": [
          "configmaps"
        ],
        "verbs": [
          "create",
          "get",
          "list",
          "watch",
          "update"
        ]
      },
      {
        "apiGroups": [
          ""
        ],
        "resources": [
          "endpoints",
          "pods",
          "services",
          "namespaces",
          "nodes",
          "secrets"
        ],
        "verbs": [
          "get",
          "list",
          "watch"
        ]
      },
      {
        "apiGroups": [
          ""
        ],
        "resources": [
          "secrets"
        ],
        "verbs": [
          "create",
          "get",
          "watch",
          "list",
          "update",
          "delete"
        ]
      },
      {
        "apiGroups": [
          "certificates.k8s.io"
        ],
        "resources": [
          "certificatesigningrequests",
          "certificatesigningrequests/approval",
          "certificatesigningrequests/status"
        ],
        "verbs": [
          "update",
          "create",
          "get",
          "delete"
        ]
      }
    ]
  },
  "istio-obj-62": {
    "apiVersion": "rbac.authorization.k8s.io/v1",
    "kind": "ClusterRoleBinding",
    "metadata": {
      "name": "istio-pilot-istio-system",
      "labels": {
        "app": "pilot",
        "release": "istio"
      }
    },
    "roleRef": {
      "apiGroup": "rbac.authorization.k8s.io",
      "kind": "ClusterRole",
      "name": "istio-pilot-istio-system"
    },
    "subjects": [
      {
        "kind": "ServiceAccount",
        "name": "istio-pilot-service-account",
        "namespace": "istio-system"
      }
    ]
  },
  "istio-obj-63": {
    "apiVersion": "v1",
    "kind": "ConfigMap",
    "metadata": {
      "namespace": "istio-system",
      "name": "pilot-envoy-config",
      "labels": {
        "release": "istio"
      }
    },
    "data": {
      "envoy.yaml.tmpl": "admin:\n  access_log_path: /dev/null\n  address:\n    socket_address:\n      address: 127.0.0.1\n      port_value: 15000\n\nstatic_resources:\n  clusters:\n  - name: in.15010\n    http2_protocol_options: {}\n    connect_timeout: 1.000s\n\n    hosts:\n    - socket_address:\n        address: 127.0.0.1\n        port_value: 15010\n\n    circuit_breakers:\n      thresholds:\n      - max_connections: 100000\n        max_pending_requests: 100000\n        max_requests: 100000\n        max_retries: 3\n\n# TODO: telemetry using EDS\n# TODO: other pilots using EDS, load balancing\n# TODO: galley using EDS\n\n  - name: out.galley.15019\n    http2_protocol_options: {}\n    connect_timeout: 1.000s\n    type: STRICT_DNS\n\n    circuit_breakers:\n      thresholds:\n        - max_connections: 100000\n          max_pending_requests: 100000\n          max_requests: 100000\n          max_retries: 3\n    hosts:\n      - socket_address:\n          address: istio-galley.istio-system\n          port_value: 15019\n    tls_context:\n      common_tls_context:\n        tls_certificates:\n        - certificate_chain:\n            filename: /etc/certs/cert-chain.pem\n          private_key:\n            filename: /etc/certs/key.pem\n        validation_context:\n          trusted_ca:\n            filename: /etc/certs/root-cert.pem\n          verify_subject_alt_name:\n          - spiffe://cluster.local/ns/istio-system/sa/istio-galley-service-account\n\n  listeners:\n  - name: \"in.15011\"\n    address:\n      socket_address:\n        address: 0.0.0.0\n        port_value: 15011\n    filter_chains:\n    - filters:\n      - name: envoy.http_connection_manager\n        #typed_config\n        #\"@type\": \"type.googleapis.com/\",\n        config:\n          codec_type: HTTP2\n          stat_prefix: \"15011\"\n          http2_protocol_options:\n            max_concurrent_streams: 1073741824\n\n          access_log:\n          - name: envoy.file_access_log\n            config:\n              path: /dev/stdout\n\n          http_filters:\n          - name: envoy.router\n\n          route_config:\n            name: \"15011\"\n\n            virtual_hosts:\n            - name: istio-pilot\n\n              domains:\n              - '*'\n\n              routes:\n              - match:\n                  prefix: /\n                route:\n                  cluster: in.15010\n                  timeout: 0.000s\n                decorator:\n                  operation: xDS\n      tls_context:\n        common_tls_context:\n          alpn_protocols:\n          - h2\n          tls_certificates:\n          - certificate_chain:\n              filename: /etc/certs/cert-chain.pem\n            private_key:\n              filename: /etc/certs/key.pem\n          validation_context:\n            trusted_ca:\n              filename: /etc/certs/root-cert.pem\n        require_client_certificate: true\n\n\n  # Manual 'whitebox' mode\n  - name: \"local.15019\"\n    address:\n      socket_address:\n        address: 127.0.0.1\n        port_value: 15019\n    filter_chains:\n      - filters:\n          - name: envoy.http_connection_manager\n            config:\n              codec_type: HTTP2\n              stat_prefix: \"15019\"\n              http2_protocol_options:\n                max_concurrent_streams: 1073741824\n\n              access_log:\n                - name: envoy.file_access_log\n                  config:\n                    path: /dev/stdout\n\n              http_filters:\n                - name: envoy.router\n\n              route_config:\n                name: \"15019\"\n\n                virtual_hosts:\n                  - name: istio-galley\n\n                    domains:\n                      - '*'\n\n                    routes:\n                      - match:\n                          prefix: /\n                        route:\n                          cluster: out.galley.15019\n                          timeout: 0.000s"
    }
  },
  "istio-obj-64": {
    "apiVersion": "v1",
    "kind": "ConfigMap",
    "metadata": {
      "name": "istio",
      "namespace": "istio-system",
      "labels": {
        "release": "istio"
      }
    },
    "data": {
      "meshNetworks": "# Network config\nnetworks: {}",
      "values.yaml": "appNamespaces: []\nautoscaleEnabled: true\nautoscaleMax: 5\nautoscaleMin: 1\nconfigMap: true\nconfigNamespace: istio-config\ncpu:\n  targetAverageUtilization: 80\nenableProtocolSniffingForInbound: false\nenableProtocolSniffingForOutbound: true\nenabled: true\nenv: {}\nimage: pilot\ningress:\n  ingressClass: istio\n  ingressControllerMode: \"OFF\"\n  ingressService: istio-ingressgateway\nkeepaliveMaxServerConnectionAge: 30m\nmeshNetworks:\n  networks: {}\nnamespace: istio-system\nnodeSelector: {}\nplugins: []\npodAnnotations: {}\npodAntiAffinityLabelSelector: []\npodAntiAffinityTermLabelSelector: []\npolicy:\n  enabled: false\nreplicaCount: 1\nresources:\n  requests:\n    cpu: 500m\n    memory: 2048Mi\nrollingMaxSurge: 100%\nrollingMaxUnavailable: 25%\ntolerations: []\ntraceSampling: 1\nuseMCP: true",
      "mesh": "# Set enableTracing to false to disable request tracing.\nenableTracing: true\n\n# Set accessLogFile to empty string to disable access log.\naccessLogFile: \"\"\n\nenableEnvoyAccessLogService: false\nmixerCheckServer: istio-policy.istio-system.svc.cluster.local:15004\nmixerReportServer: istio-telemetry.istio-system.svc.cluster.local:15004\n# policyCheckFailOpen allows traffic in cases when the mixer policy service cannot be reached.\n# Default is false which means the traffic is denied when the client is unable to connect to Mixer.\npolicyCheckFailOpen: false\n# reportBatchMaxEntries is the number of requests that are batched before telemetry data is sent to the mixer server\nreportBatchMaxEntries: 100\n# reportBatchMaxTime is the max waiting time before the telemetry data of a request is sent to the mixer server\nreportBatchMaxTime: 1s\ndisableMixerHttpReports: false\n\ndisablePolicyChecks: true\n\n\n# This is the k8s ingress service name, update if you used a different name\ningressService: \"istio-ingressgateway\"\ningressControllerMode: \"OFF\"\ningressClass: \"istio\"\n\n# The trust domain corresponds to the trust root of a system.\n# Refer to https://github.com/spiffe/spiffe/blob/master/standards/SPIFFE-ID.md#21-trust-domain\ntrustDomain: \"cluster.local\"\n\n#  The trust domain aliases represent the aliases of trust_domain.\n#  For example, if we have\n#  trustDomain: td1\n#  trustDomainAliases: [\u201ctd2\u201d, \"td3\"]\n#  Any service with the identity \"td1/ns/foo/sa/a-service-account\", \"td2/ns/foo/sa/a-service-account\",\n#  or \"td3/ns/foo/sa/a-service-account\" will be treated the same in the Istio mesh.\ntrustDomainAliases:\n\n# Set expected values when SDS is disabled\n# Unix Domain Socket through which envoy communicates with NodeAgent SDS to get\n# key/cert for mTLS. Use secret-mount files instead of SDS if set to empty.\nsdsUdsPath: \"\"\n\n# This flag is used by secret discovery service(SDS).\n# If set to true(prerequisite: https://kubernetes.io/docs/concepts/storage/volumes/#projected), Istio will inject volumes mount\n# for k8s service account JWT, so that K8s API server mounts k8s service account JWT to envoy container, which\n# will be used to generate key/cert eventually. This isn't supported for non-k8s case.\nenableSdsTokenMount: false\n\n# This flag is used by secret discovery service(SDS).\n# If set to true, envoy will fetch normal k8s service account JWT from '/var/run/secrets/kubernetes.io/serviceaccount/token'\n# (https://kubernetes.io/docs/tasks/access-application-cluster/access-cluster/#accessing-the-api-from-a-pod)\n# and pass to sds server, which will be used to request key/cert eventually.\n# this flag is ignored if enableSdsTokenMount is set.\n# This isn't supported for non-k8s case.\nsdsUseK8sSaJwt: false\n\n# If true, automatically configure client side mTLS settings to match the corresponding service's\n# server side mTLS authentication policy, when destination rule for that service does not specify\n# TLS settings.\nenableAutoMtls: false\nconfig_sources:\n- address: localhost:15019\n\noutboundTrafficPolicy:\n  mode: ALLOW_ANY\nlocalityLbSetting:\n  enabled: true\n\n# Configures DNS certificates provisioned through Chiron linked into Pilot.\n# The DNS certificate provisioning is enabled by default now so it get tested.\n# TODO (lei-tang): we'll decide whether enable it by default or not before Istio 1.4 Release.\ncertificates:\n  []\n\ndefaultConfig:\n  #\n  # TCP connection timeout between Envoy & the application, and between Envoys.\n  connectTimeout: 10s\n  #\n  ### ADVANCED SETTINGS #############\n  # Where should envoy's configuration be stored in the istio-proxy container\n  configPath: \"/etc/istio/proxy\"\n  # The pseudo service name used for Envoy.\n  serviceCluster: istio-proxy\n  # These settings that determine how long an old Envoy\n  # process should be kept alive after an occasional reload.\n  drainDuration: 45s\n  parentShutdownDuration: 1m0s\n  #\n  # Port where Envoy listens (on local host) for admin commands\n  # You can exec into the istio-proxy container in a pod and\n  # curl the admin port (curl http://localhost:15000/) to obtain\n  # diagnostic information from Envoy. See\n  # https://lyft.github.io/envoy/docs/operations/admin.html\n  # for more details\n  proxyAdminPort: 15000\n  #\n  # Set concurrency to a specific number to control the number of Proxy worker threads.\n  # If set to 0 (default), then start worker thread for each CPU thread/core.\n  concurrency: 2\n  #\n  tracing:\n    zipkin:\n      # Address of the Zipkin collector\n      address: zipkin.istio-system:9411\n  #\n  # Mutual TLS authentication between sidecars and istio control plane.\n  controlPlaneAuthPolicy: MUTUAL_TLS\n  #\n  # Address where istio Pilot service is running\n  discoveryAddress: istio-pilot.istio-system:15011"
    }
  },
  "istio-obj-65": {
    "apiVersion": "apps/v1",
    "kind": "Deployment",
    "metadata": {
      "labels": {
        "app": "pilot",
        "istio": "pilot",
        "release": "istio"
      },
      "name": "istio-pilot",
      "namespace": "istio-system"
    },
    "spec": {
      "selector": {
        "matchLabels": {
          "istio": "pilot"
        }
      },
      "strategy": {
        "rollingUpdate": {
          "maxSurge": "100%",
          "maxUnavailable": "25%"
        }
      },
      "template": {
        "metadata": {
          "annotations": {
            "sidecar.istio.io/inject": "false"
          },
          "labels": {
            "app": "pilot",
            "chart": "pilot",
            "heritage": "Tiller",
            "istio": "pilot",
            "release": "istio"
          }
        },
        "spec": {
          "affinity": {
            "nodeAffinity": {
              "preferredDuringSchedulingIgnoredDuringExecution": [
                {
                  "preference": {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "amd64"
                        ]
                      }
                    ]
                  },
                  "weight": 2
                },
                {
                  "preference": {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "ppc64le"
                        ]
                      }
                    ]
                  },
                  "weight": 2
                },
                {
                  "preference": {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "s390x"
                        ]
                      }
                    ]
                  },
                  "weight": 2
                }
              ],
              "requiredDuringSchedulingIgnoredDuringExecution": {
                "nodeSelectorTerms": [
                  {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "amd64",
                          "ppc64le",
                          "s390x"
                        ]
                      }
                    ]
                  }
                ]
              }
            }
          },
          "containers": [
            {
              "args": [
                "discovery",
                "--monitoringAddr=:15014",
                "--log_output_level=default:info",
                "--domain",
                "cluster.local",
                "--secureGrpcAddr",
                "",
                "--trust-domain=cluster.local",
                "--keepaliveMaxServerConnectionAge",
                "30m"
              ],
              "env": [
                {
                  "name": "POD_NAME",
                  "valueFrom": {
                    "fieldRef": {
                      "apiVersion": "v1",
                      "fieldPath": "metadata.name"
                    }
                  }
                },
                {
                  "name": "POD_NAMESPACE",
                  "valueFrom": {
                    "fieldRef": {
                      "apiVersion": "v1",
                      "fieldPath": "metadata.namespace"
                    }
                  }
                },
                {
                  "name": "PILOT_TRACE_SAMPLING",
                  "value": "1"
                },
                {
                  "name": "CONFIG_NAMESPACE",
                  "value": "istio-config"
                },
                {
                  "name": "PILOT_ENABLE_PROTOCOL_SNIFFING_FOR_OUTBOUND",
                  "value": "true"
                },
                {
                  "name": "PILOT_ENABLE_PROTOCOL_SNIFFING_FOR_INBOUND",
                  "value": "false"
                }
              ],
              "image": "docker.io/istio/pilot:1.4.3",
              "imagePullPolicy": "IfNotPresent",
              "name": "discovery",
              "ports": [
                {
                  "containerPort": 8080
                },
                {
                  "containerPort": 15010
                }
              ],
              "readinessProbe": {
                "httpGet": {
                  "path": "/ready",
                  "port": 8080
                },
                "initialDelaySeconds": 5,
                "periodSeconds": 30,
                "timeoutSeconds": 5
              },
              "resources": {
                "requests": {
                  "cpu": "500m",
                  "memory": "2048Mi"
                }
              },
              "volumeMounts": [
                {
                  "mountPath": "/etc/istio/config",
                  "name": "config-volume"
                }
              ]
            },
            {
              "args": [
                "proxy",
                "--domain",
                "$(POD_NAMESPACE).svc.cluster.local",
                "--serviceCluster",
                "istio-pilot",
                "--templateFile",
                "/var/lib/envoy/envoy.yaml.tmpl",
                "--controlPlaneAuthPolicy",
                "MUTUAL_TLS",
                "--trust-domain=cluster.local"
              ],
              "env": [
                {
                  "name": "POD_NAME",
                  "valueFrom": {
                    "fieldRef": {
                      "apiVersion": "v1",
                      "fieldPath": "metadata.name"
                    }
                  }
                },
                {
                  "name": "POD_NAMESPACE",
                  "valueFrom": {
                    "fieldRef": {
                      "apiVersion": "v1",
                      "fieldPath": "metadata.namespace"
                    }
                  }
                },
                {
                  "name": "INSTANCE_IP",
                  "valueFrom": {
                    "fieldRef": {
                      "apiVersion": "v1",
                      "fieldPath": "status.podIP"
                    }
                  }
                },
                {
                  "name": "SDS_ENABLED",
                  "value": "false"
                }
              ],
              "image": "docker.io/istio/proxyv2:1.4.3",
              "imagePullPolicy": "IfNotPresent",
              "name": "istio-proxy",
              "ports": [
                {
                  "containerPort": 15011
                }
              ],
              "resources": {
                "limits": {
                  "cpu": "2000m",
                  "memory": "1024Mi"
                },
                "requests": {
                  "cpu": "100m",
                  "memory": "128Mi"
                }
              },
              "volumeMounts": [
                {
                  "mountPath": "/var/lib/envoy",
                  "name": "pilot-envoy-config"
                },
                {
                  "mountPath": "/etc/certs",
                  "name": "istio-certs",
                  "readOnly": true
                }
              ]
            }
          ],
          "serviceAccountName": "istio-pilot-service-account",
          "volumes": [
            {
              "name": "istio-certs",
              "secret": {
                "optional": true,
                "secretName": "istio.istio-pilot-service-account"
              }
            },
            {
              "configMap": {
                "name": "istio"
              },
              "name": "config-volume"
            },
            {
              "configMap": {
                "name": "pilot-envoy-config"
              },
              "name": "pilot-envoy-config"
            }
          ]
        }
      }
    }
  },
  "istio-obj-67": {
    "apiVersion": "policy/v1beta1",
    "kind": "PodDisruptionBudget",
    "metadata": {
      "name": "istio-pilot",
      "namespace": "istio-system",
      "labels": {
        "app": "pilot",
        "release": "istio",
        "istio": "pilot"
      }
    },
    "spec": {
      "minAvailable": 1,
      "selector": {
        "matchLabels": {
          "app": "pilot",
          "release": "istio",
          "istio": "pilot"
        }
      }
    }
  },
  "istio-obj-68": {
    "apiVersion": "v1",
    "kind": "Service",
    "metadata": {
      "name": "istio-pilot",
      "namespace": "istio-system",
      "labels": {
        "app": "pilot",
        "release": "istio",
        "istio": "pilot"
      }
    },
    "spec": {
      "ports": [
        {
          "port": 15010,
          "name": "grpc-xds"
        },
        {
          "port": 15011,
          "name": "https-xds"
        },
        {
          "port": 8080,
          "name": "http-legacy-discovery"
        },
        {
          "port": 15014,
          "name": "http-monitoring"
        }
      ],
      "selector": {
        "istio": "pilot"
      }
    }
  },
  "istio-obj-69": {
    "apiVersion": "v1",
    "kind": "ServiceAccount",
    "metadata": {
      "name": "istio-pilot-service-account",
      "namespace": "istio-system",
      "labels": {
        "app": "pilot",
        "release": "istio"
      }
    }
  },
  "istio-obj-70": {
    "apiVersion": "autoscaling/v2beta1",
    "kind": "HorizontalPodAutoscaler",
    "metadata": {
      "labels": {
        "app": "mixer",
        "release": "istio"
      },
      "name": "istio-policy",
      "namespace": "istio-system"
    },
    "spec": {
      "maxReplicas": 5,
      "metrics": [
        {
          "resource": {
            "name": "cpu",
            "targetAverageUtilization": 80
          },
          "type": "Resource"
        }
      ],
      "minReplicas": 1,
      "scaleTargetRef": {
        "apiVersion": "apps/v1",
        "kind": "Deployment",
        "name": "istio-policy"
      }
    }
  },
  "istio-obj-71": {
    "apiVersion": "rbac.authorization.k8s.io/v1",
    "kind": "ClusterRole",
    "metadata": {
      "name": "istio-policy",
      "labels": {
        "release": "istio",
        "app": "istio-policy"
      }
    },
    "rules": [
      {
        "apiGroups": [
          "config.istio.io"
        ],
        "resources": [
          "*"
        ],
        "verbs": [
          "create",
          "get",
          "list",
          "watch",
          "patch"
        ]
      },
      {
        "apiGroups": [
          "apiextensions.k8s.io"
        ],
        "resources": [
          "customresourcedefinitions"
        ],
        "verbs": [
          "get",
          "list",
          "watch"
        ]
      },
      {
        "apiGroups": [
          ""
        ],
        "resources": [
          "configmaps",
          "endpoints",
          "pods",
          "services",
          "namespaces",
          "secrets",
          "replicationcontrollers"
        ],
        "verbs": [
          "get",
          "list",
          "watch"
        ]
      },
      {
        "apiGroups": [
          "extensions",
          "apps"
        ],
        "resources": [
          "replicasets"
        ],
        "verbs": [
          "get",
          "list",
          "watch"
        ]
      }
    ]
  },
  "istio-obj-72": {
    "apiVersion": "rbac.authorization.k8s.io/v1",
    "kind": "ClusterRoleBinding",
    "metadata": {
      "name": "istio-policy-admin-role-binding-istio-system",
      "labels": {
        "app": "istio-policy",
        "release": "istio"
      }
    },
    "roleRef": {
      "apiGroup": "rbac.authorization.k8s.io",
      "kind": "ClusterRole",
      "name": "istio-policy"
    },
    "subjects": [
      {
        "kind": "ServiceAccount",
        "name": "istio-policy-service-account",
        "namespace": "istio-system"
      }
    ]
  },
  "istio-obj-73": {
    "apiVersion": "networking.istio.io/v1alpha3",
    "kind": "DestinationRule",
    "metadata": {
      "name": "istio-policy",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-policy",
        "release": "istio"
      }
    },
    "spec": {
      "host": "istio-policy.istio-system.svc.cluster.local",
      "trafficPolicy": {
        "portLevelSettings": [
          {
            "port": {
              "number": 15004
            },
            "tls": {
              "mode": "ISTIO_MUTUAL"
            }
          },
          {
            "port": {
              "number": 9091
            },
            "tls": {
              "mode": "DISABLE"
            }
          }
        ],
        "connectionPool": {
          "http": {
            "http2MaxRequests": 10000,
            "maxRequestsPerConnection": 10000
          }
        }
      }
    }
  },
  "istio-obj-74": {
    "apiVersion": "v1",
    "kind": "ConfigMap",
    "metadata": {
      "namespace": "istio-system",
      "name": "policy-envoy-config",
      "labels": {
        "release": "istio"
      }
    },
    "data": {
      "envoy.yaml.tmpl": "admin:\n  access_log_path: /dev/null\n  address:\n    socket_address:\n      address: 127.0.0.1\n      port_value: 15000\nstats_config:\n  use_all_default_tags: false\n  stats_tags:\n  - tag_name: cluster_name\n    regex: '^cluster\\.((.+?(\\..+?\\.svc\\.cluster\\.local)?)\\.)'\n  - tag_name: tcp_prefix\n    regex: '^tcp\\.((.*?)\\.)\\w+?$'\n  - tag_name: response_code\n    regex: '_rq(_(\\d{3}))$'\n  - tag_name: response_code_class\n    regex: '_rq(_(\\dxx))$'\n  - tag_name: http_conn_manager_listener_prefix\n    regex: '^listener(?=\\.).*?\\.http\\.(((?:[_.[:digit:]]*|[_\\[\\]aAbBcCdDeEfF[:digit:]]*))\\.)'\n  - tag_name: http_conn_manager_prefix\n    regex: '^http\\.(((?:[_.[:digit:]]*|[_\\[\\]aAbBcCdDeEfF[:digit:]]*))\\.)'\n  - tag_name: listener_address\n    regex: '^listener\\.(((?:[_.[:digit:]]*|[_\\[\\]aAbBcCdDeEfF[:digit:]]*))\\.)'\n\nstatic_resources:\n  clusters:\n  - name: prometheus_stats\n    type: STATIC\n    connect_timeout: 0.250s\n    lb_policy: ROUND_ROBIN\n    hosts:\n    - socket_address:\n        protocol: TCP\n        address: 127.0.0.1\n        port_value: 15000\n\n  - circuit_breakers:\n      thresholds:\n      - max_connections: 100000\n        max_pending_requests: 100000\n        max_requests: 100000\n        max_retries: 3\n    connect_timeout: 1.000s\n    hosts:\n    - pipe:\n        path: /sock/mixer.socket\n    http2_protocol_options: {}\n    name: inbound_9092\n\n  - circuit_breakers:\n      thresholds:\n      - max_connections: 100000\n        max_pending_requests: 100000\n        max_requests: 100000\n        max_retries: 3\n    connect_timeout: 1.000s\n    hosts:\n    - socket_address:\n        address: istio-telemetry\n        port_value: 15004\n    http2_protocol_options: {}\n    name: mixer_report_server\n    tls_context:\n      common_tls_context:\n        tls_certificates:\n        - certificate_chain:\n            filename: /etc/certs/cert-chain.pem\n          private_key:\n            filename: /etc/certs/key.pem\n        validation_context:\n          trusted_ca:\n            filename: /etc/certs/root-cert.pem\n          verify_subject_alt_name:\n          - spiffe://cluster.local/ns/istio-system/sa/istio-mixer-service-account\n    type: STRICT_DNS\n    dns_lookup_family: V4_ONLY\n\n  - name: out.galley.15019\n    http2_protocol_options: {}\n    connect_timeout: 1.000s\n    type: STRICT_DNS\n\n    circuit_breakers:\n      thresholds:\n        - max_connections: 100000\n          max_pending_requests: 100000\n          max_requests: 100000\n          max_retries: 3\n    hosts:\n      - socket_address:\n          address: istio-galley.istio-system\n          port_value: 15019\n    tls_context:\n      common_tls_context:\n        tls_certificates:\n        - certificate_chain:\n            filename: /etc/certs/cert-chain.pem\n          private_key:\n            filename: /etc/certs/key.pem\n        validation_context:\n          trusted_ca:\n            filename: /etc/certs/root-cert.pem\n          verify_subject_alt_name:\n          - spiffe://cluster.local/ns/istio-system/sa/istio-galley-service-account\n\n  listeners:\n  - name: \"15090\"\n    address:\n      socket_address:\n        protocol: TCP\n        address: 0.0.0.0\n        port_value: 15090\n    filter_chains:\n    - filters:\n      - name: envoy.http_connection_manager\n        config:\n          codec_type: AUTO\n          stat_prefix: stats\n          route_config:\n            virtual_hosts:\n            - name: backend\n              domains:\n              - '*'\n              routes:\n              - match:\n                  prefix: /stats/prometheus\n                route:\n                  cluster: prometheus_stats\n          http_filters:\n          - name: envoy.router\n\n  - name: \"15004\"\n    address:\n      socket_address:\n        address: 0.0.0.0\n        port_value: 15004\n    filter_chains:\n    - filters:\n      - config:\n          codec_type: HTTP2\n          http2_protocol_options:\n            max_concurrent_streams: 1073741824\n          generate_request_id: true\n          http_filters:\n          - config:\n              default_destination_service: istio-policy.istio-system.svc.cluster.local\n              service_configs:\n                istio-policy.istio-system.svc.cluster.local:\n                  disable_check_calls: true\n{{- if .DisableReportCalls }}\n                  disable_report_calls: true\n{{- end }}\n                  mixer_attributes:\n                    attributes:\n                      destination.service.host:\n                        string_value: istio-policy.istio-system.svc.cluster.local\n                      destination.service.uid:\n                        string_value: istio://istio-system/services/istio-policy\n                      destination.service.name:\n                        string_value: istio-policy\n                      destination.service.namespace:\n                        string_value: istio-system\n                      destination.uid:\n                        string_value: kubernetes://{{ .PodName }}.istio-system\n                      destination.namespace:\n                        string_value: istio-system\n                      destination.ip:\n                        bytes_value: {{ .PodIP }}\n                      destination.port:\n                        int64_value: 15004\n                      context.reporter.kind:\n                        string_value: inbound\n                      context.reporter.uid:\n                        string_value: kubernetes://{{ .PodName }}.istio-system\n              transport:\n                check_cluster: mixer_check_server\n                report_cluster: mixer_report_server\n                attributes_for_mixer_proxy:\n                  attributes:\n                    source.uid:\n                      string_value: kubernetes://{{ .PodName }}.istio-system\n            name: mixer\n          - name: envoy.router\n          route_config:\n            name: \"15004\"\n            virtual_hosts:\n            - domains:\n              - '*'\n              name: istio-policy.istio-system.svc.cluster.local\n              routes:\n              - decorator:\n                  operation: Check\n                match:\n                  prefix: /\n                route:\n                  cluster: inbound_9092\n                  timeout: 0.000s\n          stat_prefix: \"15004\"\n        name: envoy.http_connection_manager\n      tls_context:\n        common_tls_context:\n          alpn_protocols:\n          - h2\n          tls_certificates:\n          - certificate_chain:\n              filename: /etc/certs/cert-chain.pem\n            private_key:\n              filename: /etc/certs/key.pem\n          validation_context:\n            trusted_ca:\n              filename: /etc/certs/root-cert.pem\n        require_client_certificate: true\n\n  - name: \"9091\"\n    address:\n      socket_address:\n        address: 0.0.0.0\n        port_value: 9091\n    filter_chains:\n    - filters:\n      - config:\n          codec_type: HTTP2\n          http2_protocol_options:\n            max_concurrent_streams: 1073741824\n          generate_request_id: true\n          http_filters:\n          - config:\n              default_destination_service: istio-policy.istio-system.svc.cluster.local\n              service_configs:\n                istio-policy.istio-system.svc.cluster.local:\n                  disable_check_calls: true\n{{- if .DisableReportCalls }}\n                  disable_report_calls: true\n{{- end }}\n                  mixer_attributes:\n                    attributes:\n                      destination.service.host:\n                        string_value: istio-policy.istio-system.svc.cluster.local\n                      destination.service.uid:\n                        string_value: istio://istio-system/services/istio-policy\n                      destination.service.name:\n                        string_value: istio-policy\n                      destination.service.namespace:\n                        string_value: istio-system\n                      destination.uid:\n                        string_value: kubernetes://{{ .PodName }}.istio-system\n                      destination.namespace:\n                        string_value: istio-system\n                      destination.ip:\n                        bytes_value: {{ .PodIP }}\n                      destination.port:\n                        int64_value: 9091\n                      context.reporter.kind:\n                        string_value: inbound\n                      context.reporter.uid:\n                        string_value: kubernetes://{{ .PodName }}.istio-system\n              transport:\n                check_cluster: mixer_check_server\n                report_cluster: mixer_report_server\n                attributes_for_mixer_proxy:\n                  attributes:\n                    source.uid:\n                      string_value: kubernetes://{{ .PodName }}.istio-system\n            name: mixer\n          - name: envoy.router\n          route_config:\n            name: \"9091\"\n            virtual_hosts:\n            - domains:\n              - '*'\n              name: istio-policy.istio-system.svc.cluster.local\n              routes:\n              - decorator:\n                  operation: Check\n                match:\n                  prefix: /\n                route:\n                  cluster: inbound_9092\n                  timeout: 0.000s\n          stat_prefix: \"9091\"\n        name: envoy.http_connection_manager\n    name: \"9091\"\n\n  - name: \"local.15019\"\n    address:\n      socket_address:\n        address: 127.0.0.1\n        port_value: 15019\n    filter_chains:\n      - filters:\n          - name: envoy.http_connection_manager\n            config:\n              codec_type: HTTP2\n              stat_prefix: \"15019\"\n              http2_protocol_options:\n                max_concurrent_streams: 1073741824\n\n              access_log:\n                - name: envoy.file_access_log\n                  config:\n                    path: /dev/stdout\n\n              http_filters:\n                - name: envoy.router\n\n              route_config:\n                name: \"15019\"\n\n                virtual_hosts:\n                  - name: istio-galley\n\n                    domains:\n                      - '*'\n\n                    routes:\n                      - match:\n                          prefix: /\n                        route:\n                          cluster: out.galley.15019\n                          timeout: 0.000s"
    }
  },
  "istio-obj-75": {
    "apiVersion": "apps/v1",
    "kind": "Deployment",
    "metadata": {
      "labels": {
        "app": "istio-policy",
        "istio": "mixer",
        "release": "istio"
      },
      "name": "istio-policy",
      "namespace": "istio-system"
    },
    "spec": {
      "selector": {
        "matchLabels": {
          "istio": "mixer",
          "istio-mixer-type": "policy"
        }
      },
      "strategy": {
        "rollingUpdate": {
          "maxSurge": "100%",
          "maxUnavailable": "25%"
        }
      },
      "template": {
        "metadata": {
          "annotations": {
            "sidecar.istio.io/inject": "false"
          },
          "labels": {
            "app": "policy",
            "istio": "mixer",
            "istio-mixer-type": "policy"
          }
        },
        "spec": {
          "affinity": {
            "nodeAffinity": {
              "preferredDuringSchedulingIgnoredDuringExecution": [
                {
                  "preference": {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "amd64"
                        ]
                      }
                    ]
                  },
                  "weight": 2
                },
                {
                  "preference": {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "ppc64le"
                        ]
                      }
                    ]
                  },
                  "weight": 2
                },
                {
                  "preference": {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "s390x"
                        ]
                      }
                    ]
                  },
                  "weight": 2
                }
              ],
              "requiredDuringSchedulingIgnoredDuringExecution": {
                "nodeSelectorTerms": [
                  {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "amd64",
                          "ppc64le",
                          "s390x"
                        ]
                      }
                    ]
                  }
                ]
              }
            }
          },
          "containers": [
            {
              "args": [
                "--monitoringPort=15014",
                "--address",
                "unix:///sock/mixer.socket",
                "--log_output_level=default:info",
                "--configStoreURL=mcp://localhost:15019",
                "--configDefaultNamespace=istio-system",
                "--useAdapterCRDs=false",
                "--useTemplateCRDs=false",
                "--trace_zipkin_url=http://zipkin.istio-system:9411/api/v1/spans"
              ],
              "env": [
                {
                  "name": "POD_NAMESPACE",
                  "valueFrom": {
                    "fieldRef": {
                      "apiVersion": "v1",
                      "fieldPath": "metadata.namespace"
                    }
                  }
                }
              ],
              "image": "docker.io/istio/mixer:1.4.3",
              "imagePullPolicy": "IfNotPresent",
              "livenessProbe": {
                "httpGet": {
                  "path": "/version",
                  "port": 15014
                },
                "initialDelaySeconds": 5,
                "periodSeconds": 5
              },
              "name": "mixer",
              "ports": [
                {
                  "containerPort": 9091
                },
                {
                  "containerPort": 15014
                },
                {
                  "containerPort": 42422
                }
              ],
              "resources": {
                "requests": {
                  "cpu": "10m"
                }
              },
              "volumeMounts": [
                {
                  "mountPath": "/etc/certs",
                  "name": "istio-certs",
                  "readOnly": true
                },
                {
                  "mountPath": "/sock",
                  "name": "uds-socket"
                },
                {
                  "mountPath": "/var/run/secrets/istio.io/policy/adapter",
                  "name": "policy-adapter-secret",
                  "readOnly": true
                }
              ]
            },
            {
              "args": [
                "proxy",
                "--domain",
                "$(POD_NAMESPACE).svc.cluster.local",
                "--serviceCluster",
                "istio-policy",
                "--templateFile",
                "/var/lib/envoy/envoy.yaml.tmpl",
                "--controlPlaneAuthPolicy",
                "MUTUAL_TLS",
                "--trust-domain=cluster.local"
              ],
              "env": [
                {
                  "name": "POD_NAME",
                  "valueFrom": {
                    "fieldRef": {
                      "apiVersion": "v1",
                      "fieldPath": "metadata.name"
                    }
                  }
                },
                {
                  "name": "POD_NAMESPACE",
                  "valueFrom": {
                    "fieldRef": {
                      "apiVersion": "v1",
                      "fieldPath": "metadata.namespace"
                    }
                  }
                },
                {
                  "name": "INSTANCE_IP",
                  "valueFrom": {
                    "fieldRef": {
                      "apiVersion": "v1",
                      "fieldPath": "status.podIP"
                    }
                  }
                },
                {
                  "name": "SDS_ENABLED",
                  "value": "false"
                }
              ],
              "image": "docker.io/istio/proxyv2:1.4.3",
              "imagePullPolicy": "IfNotPresent",
              "name": "istio-proxy",
              "ports": [
                {
                  "containerPort": 15004
                },
                {
                  "containerPort": 15090,
                  "name": "http-envoy-prom",
                  "protocol": "TCP"
                }
              ],
              "resources": {
                "limits": {
                  "cpu": "2000m",
                  "memory": "1024Mi"
                },
                "requests": {
                  "cpu": "100m",
                  "memory": "128Mi"
                }
              },
              "volumeMounts": [
                {
                  "mountPath": "/var/lib/envoy",
                  "name": "policy-envoy-config"
                },
                {
                  "mountPath": "/etc/certs",
                  "name": "istio-certs",
                  "readOnly": true
                },
                {
                  "mountPath": "/sock",
                  "name": "uds-socket"
                }
              ]
            }
          ],
          "serviceAccountName": "istio-policy-service-account",
          "volumes": [
            {
              "name": "istio-certs",
              "secret": {
                "optional": true,
                "secretName": "istio.istio-policy-service-account"
              }
            },
            {
              "emptyDir": {},
              "name": "uds-socket"
            },
            {
              "name": "policy-adapter-secret",
              "secret": {
                "optional": true,
                "secretName": "policy-adapter-secret"
              }
            },
            {
              "configMap": {
                "name": "policy-envoy-config"
              },
              "name": "policy-envoy-config"
            }
          ]
        }
      }
    }
  },
  "istio-obj-76": {
    "apiVersion": "policy/v1beta1",
    "kind": "PodDisruptionBudget",
    "metadata": {
      "name": "istio-policy",
      "namespace": "istio-system",
      "labels": {
        "app": "policy",
        "release": "istio",
        "istio": "mixer",
        "istio-mixer-type": "policy"
      }
    },
    "spec": {
      "minAvailable": 1,
      "selector": {
        "matchLabels": {
          "app": "policy",
          "istio": "mixer",
          "istio-mixer-type": "policy"
        }
      }
    }
  },
  "istio-obj-77": {
    "apiVersion": "v1",
    "kind": "Service",
    "metadata": {
      "name": "istio-policy",
      "namespace": "istio-system",
      "labels": {
        "app": "mixer",
        "istio": "mixer",
        "release": "istio"
      }
    },
    "spec": {
      "ports": [
        {
          "name": "grpc-mixer",
          "port": 9091
        },
        {
          "name": "grpc-mixer-mtls",
          "port": 15004
        },
        {
          "name": "http-policy-monitoring",
          "port": 15014
        }
      ],
      "selector": {
        "istio": "mixer",
        "istio-mixer-type": "policy"
      }
    }
  },
  "istio-obj-78": {
    "apiVersion": "v1",
    "kind": "ServiceAccount",
    "metadata": {
      "name": "istio-policy-service-account",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-policy",
        "release": "istio"
      }
    }
  },
  "istio-obj-79": {
    "apiVersion": "autoscaling/v2beta1",
    "kind": "HorizontalPodAutoscaler",
    "metadata": {
      "labels": {
        "app": "mixer",
        "release": "istio"
      },
      "name": "istio-telemetry",
      "namespace": "istio-system"
    },
    "spec": {
      "maxReplicas": 5,
      "metrics": [
        {
          "resource": {
            "name": "cpu",
            "targetAverageUtilization": 80
          },
          "type": "Resource"
        }
      ],
      "minReplicas": 1,
      "scaleTargetRef": {
        "apiVersion": "apps/v1",
        "kind": "Deployment",
        "name": "istio-telemetry"
      }
    }
  },
  "istio-obj-80": {
    "apiVersion": "rbac.authorization.k8s.io/v1",
    "kind": "ClusterRole",
    "metadata": {
      "name": "istio-mixer-istio-system",
      "labels": {
        "app": "istio-telemetry",
        "release": "istio"
      }
    },
    "rules": [
      {
        "apiGroups": [
          "config.istio.io"
        ],
        "resources": [
          "*"
        ],
        "verbs": [
          "create",
          "get",
          "list",
          "watch",
          "patch"
        ]
      },
      {
        "apiGroups": [
          "apiextensions.k8s.io"
        ],
        "resources": [
          "customresourcedefinitions"
        ],
        "verbs": [
          "get",
          "list",
          "watch"
        ]
      },
      {
        "apiGroups": [
          ""
        ],
        "resources": [
          "configmaps",
          "endpoints",
          "pods",
          "services",
          "namespaces",
          "secrets",
          "replicationcontrollers"
        ],
        "verbs": [
          "get",
          "list",
          "watch"
        ]
      },
      {
        "apiGroups": [
          "extensions",
          "apps"
        ],
        "resources": [
          "replicasets"
        ],
        "verbs": [
          "get",
          "list",
          "watch"
        ]
      }
    ]
  },
  "istio-obj-81": {
    "apiVersion": "rbac.authorization.k8s.io/v1",
    "kind": "ClusterRoleBinding",
    "metadata": {
      "name": "istio-mixer-admin-role-binding-istio-system",
      "labels": {
        "app": "istio-telemetry",
        "release": "istio"
      }
    },
    "roleRef": {
      "apiGroup": "rbac.authorization.k8s.io",
      "kind": "ClusterRole",
      "name": "istio-mixer-istio-system"
    },
    "subjects": [
      {
        "kind": "ServiceAccount",
        "name": "istio-mixer-service-account",
        "namespace": "istio-system"
      }
    ]
  },
  "istio-obj-82": {
    "apiVersion": "config.istio.io/v1alpha2",
    "kind": "attributemanifest",
    "metadata": {
      "name": "istioproxy",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-telemetry",
        "release": "istio"
      }
    },
    "spec": {
      "attributes": {
        "origin.ip": {
          "valueType": "IP_ADDRESS"
        },
        "origin.uid": {
          "valueType": "STRING"
        },
        "origin.user": {
          "valueType": "STRING"
        },
        "request.headers": {
          "valueType": "STRING_MAP"
        },
        "request.id": {
          "valueType": "STRING"
        },
        "request.host": {
          "valueType": "STRING"
        },
        "request.method": {
          "valueType": "STRING"
        },
        "request.path": {
          "valueType": "STRING"
        },
        "request.url_path": {
          "valueType": "STRING"
        },
        "request.query_params": {
          "valueType": "STRING_MAP"
        },
        "request.reason": {
          "valueType": "STRING"
        },
        "request.referer": {
          "valueType": "STRING"
        },
        "request.scheme": {
          "valueType": "STRING"
        },
        "request.total_size": {
          "valueType": "INT64"
        },
        "request.size": {
          "valueType": "INT64"
        },
        "request.time": {
          "valueType": "TIMESTAMP"
        },
        "request.useragent": {
          "valueType": "STRING"
        },
        "response.code": {
          "valueType": "INT64"
        },
        "response.duration": {
          "valueType": "DURATION"
        },
        "response.headers": {
          "valueType": "STRING_MAP"
        },
        "response.total_size": {
          "valueType": "INT64"
        },
        "response.size": {
          "valueType": "INT64"
        },
        "response.time": {
          "valueType": "TIMESTAMP"
        },
        "response.grpc_status": {
          "valueType": "STRING"
        },
        "response.grpc_message": {
          "valueType": "STRING"
        },
        "source.uid": {
          "valueType": "STRING"
        },
        "source.user": {
          "valueType": "STRING"
        },
        "source.principal": {
          "valueType": "STRING"
        },
        "destination.uid": {
          "valueType": "STRING"
        },
        "destination.principal": {
          "valueType": "STRING"
        },
        "destination.port": {
          "valueType": "INT64"
        },
        "connection.event": {
          "valueType": "STRING"
        },
        "connection.id": {
          "valueType": "STRING"
        },
        "connection.received.bytes": {
          "valueType": "INT64"
        },
        "connection.received.bytes_total": {
          "valueType": "INT64"
        },
        "connection.sent.bytes": {
          "valueType": "INT64"
        },
        "connection.sent.bytes_total": {
          "valueType": "INT64"
        },
        "connection.duration": {
          "valueType": "DURATION"
        },
        "connection.mtls": {
          "valueType": "BOOL"
        },
        "connection.requested_server_name": {
          "valueType": "STRING"
        },
        "context.protocol": {
          "valueType": "STRING"
        },
        "context.proxy_error_code": {
          "valueType": "STRING"
        },
        "context.timestamp": {
          "valueType": "TIMESTAMP"
        },
        "context.time": {
          "valueType": "TIMESTAMP"
        },
        "context.reporter.local": {
          "valueType": "BOOL"
        },
        "context.reporter.kind": {
          "valueType": "STRING"
        },
        "context.reporter.uid": {
          "valueType": "STRING"
        },
        "context.proxy_version": {
          "valueType": "STRING"
        },
        "api.service": {
          "valueType": "STRING"
        },
        "api.version": {
          "valueType": "STRING"
        },
        "api.operation": {
          "valueType": "STRING"
        },
        "api.protocol": {
          "valueType": "STRING"
        },
        "request.auth.principal": {
          "valueType": "STRING"
        },
        "request.auth.audiences": {
          "valueType": "STRING"
        },
        "request.auth.presenter": {
          "valueType": "STRING"
        },
        "request.auth.claims": {
          "valueType": "STRING_MAP"
        },
        "request.auth.raw_claims": {
          "valueType": "STRING"
        },
        "request.api_key": {
          "valueType": "STRING"
        },
        "rbac.permissive.response_code": {
          "valueType": "STRING"
        },
        "rbac.permissive.effective_policy_id": {
          "valueType": "STRING"
        },
        "check.error_code": {
          "valueType": "INT64"
        },
        "check.error_message": {
          "valueType": "STRING"
        },
        "check.cache_hit": {
          "valueType": "BOOL"
        },
        "quota.cache_hit": {
          "valueType": "BOOL"
        }
      }
    }
  },
  "istio-obj-83": {
    "apiVersion": "config.istio.io/v1alpha2",
    "kind": "attributemanifest",
    "metadata": {
      "name": "kubernetes",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-telemetry",
        "release": "istio"
      }
    },
    "spec": {
      "attributes": {
        "source.ip": {
          "valueType": "IP_ADDRESS"
        },
        "source.labels": {
          "valueType": "STRING_MAP"
        },
        "source.metadata": {
          "valueType": "STRING_MAP"
        },
        "source.name": {
          "valueType": "STRING"
        },
        "source.namespace": {
          "valueType": "STRING"
        },
        "source.owner": {
          "valueType": "STRING"
        },
        "source.serviceAccount": {
          "valueType": "STRING"
        },
        "source.services": {
          "valueType": "STRING"
        },
        "source.workload.uid": {
          "valueType": "STRING"
        },
        "source.workload.name": {
          "valueType": "STRING"
        },
        "source.workload.namespace": {
          "valueType": "STRING"
        },
        "destination.ip": {
          "valueType": "IP_ADDRESS"
        },
        "destination.labels": {
          "valueType": "STRING_MAP"
        },
        "destination.metadata": {
          "valueType": "STRING_MAP"
        },
        "destination.owner": {
          "valueType": "STRING"
        },
        "destination.name": {
          "valueType": "STRING"
        },
        "destination.container.name": {
          "valueType": "STRING"
        },
        "destination.namespace": {
          "valueType": "STRING"
        },
        "destination.service.uid": {
          "valueType": "STRING"
        },
        "destination.service.name": {
          "valueType": "STRING"
        },
        "destination.service.namespace": {
          "valueType": "STRING"
        },
        "destination.service.host": {
          "valueType": "STRING"
        },
        "destination.serviceAccount": {
          "valueType": "STRING"
        },
        "destination.workload.uid": {
          "valueType": "STRING"
        },
        "destination.workload.name": {
          "valueType": "STRING"
        },
        "destination.workload.namespace": {
          "valueType": "STRING"
        }
      }
    }
  },
  "istio-obj-84": {
    "apiVersion": "config.istio.io/v1alpha2",
    "kind": "instance",
    "metadata": {
      "name": "requestcount",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-telemetry",
        "release": "istio"
      }
    },
    "spec": {
      "compiledTemplate": "metric",
      "params": {
        "value": "1",
        "dimensions": {
          "reporter": "conditional((context.reporter.kind | \"inbound\") == \"outbound\", \"source\", \"destination\")",
          "source_workload": "source.workload.name | \"unknown\"",
          "source_workload_namespace": "source.workload.namespace | \"unknown\"",
          "source_principal": "source.principal | \"unknown\"",
          "source_app": "source.labels[\"app\"] | \"unknown\"",
          "source_version": "source.labels[\"version\"] | \"unknown\"",
          "destination_workload": "destination.workload.name | \"unknown\"",
          "destination_workload_namespace": "destination.workload.namespace | \"unknown\"",
          "destination_principal": "destination.principal | \"unknown\"",
          "destination_app": "destination.labels[\"app\"] | \"unknown\"",
          "destination_version": "destination.labels[\"version\"] | \"unknown\"",
          "destination_service": "destination.service.host | conditional((destination.service.name | \"unknown\") == \"unknown\", \"unknown\", request.host)",
          "destination_service_name": "destination.service.name | \"unknown\"",
          "destination_service_namespace": "destination.service.namespace | \"unknown\"",
          "request_protocol": "api.protocol | context.protocol | \"unknown\"",
          "response_code": "response.code | 200",
          "response_flags": "context.proxy_error_code | \"-\"",
          "permissive_response_code": "rbac.permissive.response_code | \"none\"",
          "permissive_response_policyid": "rbac.permissive.effective_policy_id | \"none\"",
          "connection_security_policy": "conditional((context.reporter.kind | \"inbound\") == \"outbound\", \"unknown\", conditional(connection.mtls | false, \"mutual_tls\", \"none\"))"
        },
        "monitored_resource_type": "\"UNSPECIFIED\""
      }
    }
  },
  "istio-obj-85": {
    "apiVersion": "config.istio.io/v1alpha2",
    "kind": "instance",
    "metadata": {
      "name": "requestduration",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-telemetry",
        "release": "istio"
      }
    },
    "spec": {
      "compiledTemplate": "metric",
      "params": {
        "value": "response.duration | \"0ms\"",
        "dimensions": {
          "reporter": "conditional((context.reporter.kind | \"inbound\") == \"outbound\", \"source\", \"destination\")",
          "source_workload": "source.workload.name | \"unknown\"",
          "source_workload_namespace": "source.workload.namespace | \"unknown\"",
          "source_principal": "source.principal | \"unknown\"",
          "source_app": "source.labels[\"app\"] | \"unknown\"",
          "source_version": "source.labels[\"version\"] | \"unknown\"",
          "destination_workload": "destination.workload.name | \"unknown\"",
          "destination_workload_namespace": "destination.workload.namespace | \"unknown\"",
          "destination_principal": "destination.principal | \"unknown\"",
          "destination_app": "destination.labels[\"app\"] | \"unknown\"",
          "destination_version": "destination.labels[\"version\"] | \"unknown\"",
          "destination_service": "destination.service.host | conditional((destination.service.name | \"unknown\") == \"unknown\", \"unknown\", request.host)",
          "destination_service_name": "destination.service.name | \"unknown\"",
          "destination_service_namespace": "destination.service.namespace | \"unknown\"",
          "request_protocol": "api.protocol | context.protocol | \"unknown\"",
          "response_code": "response.code | 200",
          "response_flags": "context.proxy_error_code | \"-\"",
          "permissive_response_code": "rbac.permissive.response_code | \"none\"",
          "permissive_response_policyid": "rbac.permissive.effective_policy_id | \"none\"",
          "connection_security_policy": "conditional((context.reporter.kind | \"inbound\") == \"outbound\", \"unknown\", conditional(connection.mtls | false, \"mutual_tls\", \"none\"))"
        },
        "monitored_resource_type": "\"UNSPECIFIED\""
      }
    }
  },
  "istio-obj-86": {
    "apiVersion": "config.istio.io/v1alpha2",
    "kind": "instance",
    "metadata": {
      "name": "requestsize",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-telemetry",
        "release": "istio"
      }
    },
    "spec": {
      "compiledTemplate": "metric",
      "params": {
        "value": "request.size | 0",
        "dimensions": {
          "reporter": "conditional((context.reporter.kind | \"inbound\") == \"outbound\", \"source\", \"destination\")",
          "source_workload": "source.workload.name | \"unknown\"",
          "source_workload_namespace": "source.workload.namespace | \"unknown\"",
          "source_principal": "source.principal | \"unknown\"",
          "source_app": "source.labels[\"app\"] | \"unknown\"",
          "source_version": "source.labels[\"version\"] | \"unknown\"",
          "destination_workload": "destination.workload.name | \"unknown\"",
          "destination_workload_namespace": "destination.workload.namespace | \"unknown\"",
          "destination_principal": "destination.principal | \"unknown\"",
          "destination_app": "destination.labels[\"app\"] | \"unknown\"",
          "destination_version": "destination.labels[\"version\"] | \"unknown\"",
          "destination_service": "destination.service.host | conditional((destination.service.name | \"unknown\") == \"unknown\", \"unknown\", request.host)",
          "destination_service_name": "destination.service.name | \"unknown\"",
          "destination_service_namespace": "destination.service.namespace | \"unknown\"",
          "request_protocol": "api.protocol | context.protocol | \"unknown\"",
          "response_code": "response.code | 200",
          "response_flags": "context.proxy_error_code | \"-\"",
          "permissive_response_code": "rbac.permissive.response_code | \"none\"",
          "permissive_response_policyid": "rbac.permissive.effective_policy_id | \"none\"",
          "connection_security_policy": "conditional((context.reporter.kind | \"inbound\") == \"outbound\", \"unknown\", conditional(connection.mtls | false, \"mutual_tls\", \"none\"))"
        },
        "monitored_resource_type": "\"UNSPECIFIED\""
      }
    }
  },
  "istio-obj-87": {
    "apiVersion": "config.istio.io/v1alpha2",
    "kind": "instance",
    "metadata": {
      "name": "responsesize",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-telemetry",
        "release": "istio"
      }
    },
    "spec": {
      "compiledTemplate": "metric",
      "params": {
        "value": "response.size | 0",
        "dimensions": {
          "reporter": "conditional((context.reporter.kind | \"inbound\") == \"outbound\", \"source\", \"destination\")",
          "source_workload": "source.workload.name | \"unknown\"",
          "source_workload_namespace": "source.workload.namespace | \"unknown\"",
          "source_principal": "source.principal | \"unknown\"",
          "source_app": "source.labels[\"app\"] | \"unknown\"",
          "source_version": "source.labels[\"version\"] | \"unknown\"",
          "destination_workload": "destination.workload.name | \"unknown\"",
          "destination_workload_namespace": "destination.workload.namespace | \"unknown\"",
          "destination_principal": "destination.principal | \"unknown\"",
          "destination_app": "destination.labels[\"app\"] | \"unknown\"",
          "destination_version": "destination.labels[\"version\"] | \"unknown\"",
          "destination_service": "destination.service.host | conditional((destination.service.name | \"unknown\") == \"unknown\", \"unknown\", request.host)",
          "destination_service_name": "destination.service.name | \"unknown\"",
          "destination_service_namespace": "destination.service.namespace | \"unknown\"",
          "request_protocol": "api.protocol | context.protocol | \"unknown\"",
          "response_code": "response.code | 200",
          "response_flags": "context.proxy_error_code | \"-\"",
          "permissive_response_code": "rbac.permissive.response_code | \"none\"",
          "permissive_response_policyid": "rbac.permissive.effective_policy_id | \"none\"",
          "connection_security_policy": "conditional((context.reporter.kind | \"inbound\") == \"outbound\", \"unknown\", conditional(connection.mtls | false, \"mutual_tls\", \"none\"))"
        },
        "monitored_resource_type": "\"UNSPECIFIED\""
      }
    }
  },
  "istio-obj-88": {
    "apiVersion": "config.istio.io/v1alpha2",
    "kind": "instance",
    "metadata": {
      "name": "tcpbytesent",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-telemetry",
        "release": "istio"
      }
    },
    "spec": {
      "compiledTemplate": "metric",
      "params": {
        "value": "connection.sent.bytes | 0",
        "dimensions": {
          "reporter": "conditional((context.reporter.kind | \"inbound\") == \"outbound\", \"source\", \"destination\")",
          "source_workload": "source.workload.name | \"unknown\"",
          "source_workload_namespace": "source.workload.namespace | \"unknown\"",
          "source_principal": "source.principal | \"unknown\"",
          "source_app": "source.labels[\"app\"] | \"unknown\"",
          "source_version": "source.labels[\"version\"] | \"unknown\"",
          "destination_workload": "destination.workload.name | \"unknown\"",
          "destination_workload_namespace": "destination.workload.namespace | \"unknown\"",
          "destination_principal": "destination.principal | \"unknown\"",
          "destination_app": "destination.labels[\"app\"] | \"unknown\"",
          "destination_version": "destination.labels[\"version\"] | \"unknown\"",
          "destination_service": "destination.service.host | \"unknown\"",
          "destination_service_name": "destination.service.name | \"unknown\"",
          "destination_service_namespace": "destination.service.namespace | \"unknown\"",
          "connection_security_policy": "conditional((context.reporter.kind | \"inbound\") == \"outbound\", \"unknown\", conditional(connection.mtls | false, \"mutual_tls\", \"none\"))",
          "response_flags": "context.proxy_error_code | \"-\""
        },
        "monitored_resource_type": "\"UNSPECIFIED\""
      }
    }
  },
  "istio-obj-89": {
    "apiVersion": "config.istio.io/v1alpha2",
    "kind": "instance",
    "metadata": {
      "name": "tcpbytereceived",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-telemetry",
        "release": "istio"
      }
    },
    "spec": {
      "compiledTemplate": "metric",
      "params": {
        "value": "connection.received.bytes | 0",
        "dimensions": {
          "reporter": "conditional((context.reporter.kind | \"inbound\") == \"outbound\", \"source\", \"destination\")",
          "source_workload": "source.workload.name | \"unknown\"",
          "source_workload_namespace": "source.workload.namespace | \"unknown\"",
          "source_principal": "source.principal | \"unknown\"",
          "source_app": "source.labels[\"app\"] | \"unknown\"",
          "source_version": "source.labels[\"version\"] | \"unknown\"",
          "destination_workload": "destination.workload.name | \"unknown\"",
          "destination_workload_namespace": "destination.workload.namespace | \"unknown\"",
          "destination_principal": "destination.principal | \"unknown\"",
          "destination_app": "destination.labels[\"app\"] | \"unknown\"",
          "destination_version": "destination.labels[\"version\"] | \"unknown\"",
          "destination_service": "destination.service.host | \"unknown\"",
          "destination_service_name": "destination.service.name | \"unknown\"",
          "destination_service_namespace": "destination.service.namespace | \"unknown\"",
          "connection_security_policy": "conditional((context.reporter.kind | \"inbound\") == \"outbound\", \"unknown\", conditional(connection.mtls | false, \"mutual_tls\", \"none\"))",
          "response_flags": "context.proxy_error_code | \"-\""
        },
        "monitored_resource_type": "\"UNSPECIFIED\""
      }
    }
  },
  "istio-obj-90": {
    "apiVersion": "config.istio.io/v1alpha2",
    "kind": "instance",
    "metadata": {
      "name": "tcpconnectionsopened",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-telemetry",
        "release": "istio"
      }
    },
    "spec": {
      "compiledTemplate": "metric",
      "params": {
        "value": "1",
        "dimensions": {
          "reporter": "conditional((context.reporter.kind | \"inbound\") == \"outbound\", \"source\", \"destination\")",
          "source_workload": "source.workload.name | \"unknown\"",
          "source_workload_namespace": "source.workload.namespace | \"unknown\"",
          "source_principal": "source.principal | \"unknown\"",
          "source_app": "source.labels[\"app\"] | \"unknown\"",
          "source_version": "source.labels[\"version\"] | \"unknown\"",
          "destination_workload": "destination.workload.name | \"unknown\"",
          "destination_workload_namespace": "destination.workload.namespace | \"unknown\"",
          "destination_principal": "destination.principal | \"unknown\"",
          "destination_app": "destination.labels[\"app\"] | \"unknown\"",
          "destination_version": "destination.labels[\"version\"] | \"unknown\"",
          "destination_service": "destination.service.host | \"unknown\"",
          "destination_service_name": "destination.service.name | \"unknown\"",
          "destination_service_namespace": "destination.service.namespace | \"unknown\"",
          "connection_security_policy": "conditional((context.reporter.kind | \"inbound\") == \"outbound\", \"unknown\", conditional(connection.mtls | false, \"mutual_tls\", \"none\"))",
          "response_flags": "context.proxy_error_code | \"-\""
        },
        "monitored_resource_type": "\"UNSPECIFIED\""
      }
    }
  },
  "istio-obj-91": {
    "apiVersion": "config.istio.io/v1alpha2",
    "kind": "instance",
    "metadata": {
      "name": "tcpconnectionsclosed",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-telemetry",
        "release": "istio"
      }
    },
    "spec": {
      "compiledTemplate": "metric",
      "params": {
        "value": "1",
        "dimensions": {
          "reporter": "conditional((context.reporter.kind | \"inbound\") == \"outbound\", \"source\", \"destination\")",
          "source_workload": "source.workload.name | \"unknown\"",
          "source_workload_namespace": "source.workload.namespace | \"unknown\"",
          "source_principal": "source.principal | \"unknown\"",
          "source_app": "source.labels[\"app\"] | \"unknown\"",
          "source_version": "source.labels[\"version\"] | \"unknown\"",
          "destination_workload": "destination.workload.name | \"unknown\"",
          "destination_workload_namespace": "destination.workload.namespace | \"unknown\"",
          "destination_principal": "destination.principal | \"unknown\"",
          "destination_app": "destination.labels[\"app\"] | \"unknown\"",
          "destination_version": "destination.labels[\"version\"] | \"unknown\"",
          "destination_service": "destination.service.host | \"unknown\"",
          "destination_service_name": "destination.service.name | \"unknown\"",
          "destination_service_namespace": "destination.service.namespace | \"unknown\"",
          "connection_security_policy": "conditional((context.reporter.kind | \"inbound\") == \"outbound\", \"unknown\", conditional(connection.mtls | false, \"mutual_tls\", \"none\"))",
          "response_flags": "context.proxy_error_code | \"-\""
        },
        "monitored_resource_type": "\"UNSPECIFIED\""
      }
    }
  },
  "istio-obj-92": {
    "apiVersion": "config.istio.io/v1alpha2",
    "kind": "handler",
    "metadata": {
      "name": "prometheus",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-telemetry",
        "release": "istio"
      }
    },
    "spec": {
      "compiledAdapter": "prometheus",
      "params": {
        "metricsExpirationPolicy": {
          "metricsExpiryDuration": "10m"
        },
        "metrics": [
          {
            "name": "requests_total",
            "instance_name": "requestcount.instance.istio-system",
            "kind": "COUNTER",
            "label_names": [
              "reporter",
              "source_app",
              "source_principal",
              "source_workload",
              "source_workload_namespace",
              "source_version",
              "destination_app",
              "destination_principal",
              "destination_workload",
              "destination_workload_namespace",
              "destination_version",
              "destination_service",
              "destination_service_name",
              "destination_service_namespace",
              "request_protocol",
              "response_code",
              "response_flags",
              "permissive_response_code",
              "permissive_response_policyid",
              "connection_security_policy"
            ]
          },
          {
            "name": "request_duration_seconds",
            "instance_name": "requestduration.instance.istio-system",
            "kind": "DISTRIBUTION",
            "label_names": [
              "reporter",
              "source_app",
              "source_principal",
              "source_workload",
              "source_workload_namespace",
              "source_version",
              "destination_app",
              "destination_principal",
              "destination_workload",
              "destination_workload_namespace",
              "destination_version",
              "destination_service",
              "destination_service_name",
              "destination_service_namespace",
              "request_protocol",
              "response_code",
              "response_flags",
              "permissive_response_code",
              "permissive_response_policyid",
              "connection_security_policy"
            ],
            "buckets": {
              "explicit_buckets": {
                "bounds": [
                  0.005,
                  0.01,
                  0.025,
                  0.05,
                  0.1,
                  0.25,
                  0.5,
                  1,
                  2.5,
                  5,
                  10
                ]
              }
            }
          },
          {
            "name": "request_bytes",
            "instance_name": "requestsize.instance.istio-system",
            "kind": "DISTRIBUTION",
            "label_names": [
              "reporter",
              "source_app",
              "source_principal",
              "source_workload",
              "source_workload_namespace",
              "source_version",
              "destination_app",
              "destination_principal",
              "destination_workload",
              "destination_workload_namespace",
              "destination_version",
              "destination_service",
              "destination_service_name",
              "destination_service_namespace",
              "request_protocol",
              "response_code",
              "response_flags",
              "permissive_response_code",
              "permissive_response_policyid",
              "connection_security_policy"
            ],
            "buckets": {
              "exponentialBuckets": {
                "numFiniteBuckets": 8,
                "scale": 1,
                "growthFactor": 10
              }
            }
          },
          {
            "name": "response_bytes",
            "instance_name": "responsesize.instance.istio-system",
            "kind": "DISTRIBUTION",
            "label_names": [
              "reporter",
              "source_app",
              "source_principal",
              "source_workload",
              "source_workload_namespace",
              "source_version",
              "destination_app",
              "destination_principal",
              "destination_workload",
              "destination_workload_namespace",
              "destination_version",
              "destination_service",
              "destination_service_name",
              "destination_service_namespace",
              "request_protocol",
              "response_code",
              "response_flags",
              "permissive_response_code",
              "permissive_response_policyid",
              "connection_security_policy"
            ],
            "buckets": {
              "exponentialBuckets": {
                "numFiniteBuckets": 8,
                "scale": 1,
                "growthFactor": 10
              }
            }
          },
          {
            "name": "tcp_sent_bytes_total",
            "instance_name": "tcpbytesent.instance.istio-system",
            "kind": "COUNTER",
            "label_names": [
              "reporter",
              "source_app",
              "source_principal",
              "source_workload",
              "source_workload_namespace",
              "source_version",
              "destination_app",
              "destination_principal",
              "destination_workload",
              "destination_workload_namespace",
              "destination_version",
              "destination_service",
              "destination_service_name",
              "destination_service_namespace",
              "connection_security_policy",
              "response_flags"
            ]
          },
          {
            "name": "tcp_received_bytes_total",
            "instance_name": "tcpbytereceived.instance.istio-system",
            "kind": "COUNTER",
            "label_names": [
              "reporter",
              "source_app",
              "source_principal",
              "source_workload",
              "source_workload_namespace",
              "source_version",
              "destination_app",
              "destination_principal",
              "destination_workload",
              "destination_workload_namespace",
              "destination_version",
              "destination_service",
              "destination_service_name",
              "destination_service_namespace",
              "connection_security_policy",
              "response_flags"
            ]
          },
          {
            "name": "tcp_connections_opened_total",
            "instance_name": "tcpconnectionsopened.instance.istio-system",
            "kind": "COUNTER",
            "label_names": [
              "reporter",
              "source_app",
              "source_principal",
              "source_workload",
              "source_workload_namespace",
              "source_version",
              "destination_app",
              "destination_principal",
              "destination_workload",
              "destination_workload_namespace",
              "destination_version",
              "destination_service",
              "destination_service_name",
              "destination_service_namespace",
              "connection_security_policy",
              "response_flags"
            ]
          },
          {
            "name": "tcp_connections_closed_total",
            "instance_name": "tcpconnectionsclosed.instance.istio-system",
            "kind": "COUNTER",
            "label_names": [
              "reporter",
              "source_app",
              "source_principal",
              "source_workload",
              "source_workload_namespace",
              "source_version",
              "destination_app",
              "destination_principal",
              "destination_workload",
              "destination_workload_namespace",
              "destination_version",
              "destination_service",
              "destination_service_name",
              "destination_service_namespace",
              "connection_security_policy",
              "response_flags"
            ]
          }
        ]
      }
    }
  },
  "istio-obj-93": {
    "apiVersion": "config.istio.io/v1alpha2",
    "kind": "rule",
    "metadata": {
      "name": "promhttp",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-telemetry",
        "release": "istio"
      }
    },
    "spec": {
      "match": "(context.protocol == \"http\" || context.protocol == \"grpc\") && (match((request.useragent | \"-\"), \"kube-probe*\") == false) && (match((request.useragent | \"-\"), \"Prometheus*\") == false)",
      "actions": [
        {
          "handler": "prometheus",
          "instances": [
            "requestcount",
            "requestduration",
            "requestsize",
            "responsesize"
          ]
        }
      ]
    }
  },
  "istio-obj-94": {
    "apiVersion": "config.istio.io/v1alpha2",
    "kind": "rule",
    "metadata": {
      "name": "promtcp",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-telemetry",
        "release": "istio"
      }
    },
    "spec": {
      "match": "context.protocol == \"tcp\"",
      "actions": [
        {
          "handler": "prometheus",
          "instances": [
            "tcpbytesent",
            "tcpbytereceived"
          ]
        }
      ]
    }
  },
  "istio-obj-95": {
    "apiVersion": "config.istio.io/v1alpha2",
    "kind": "rule",
    "metadata": {
      "name": "promtcpconnectionopen",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-telemetry",
        "release": "istio"
      }
    },
    "spec": {
      "match": "context.protocol == \"tcp\" && ((connection.event | \"na\") == \"open\")",
      "actions": [
        {
          "handler": "prometheus",
          "instances": [
            "tcpconnectionsopened"
          ]
        }
      ]
    }
  },
  "istio-obj-96": {
    "apiVersion": "config.istio.io/v1alpha2",
    "kind": "rule",
    "metadata": {
      "name": "promtcpconnectionclosed",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-telemetry",
        "release": "istio"
      }
    },
    "spec": {
      "match": "context.protocol == \"tcp\" && ((connection.event | \"na\") == \"close\")",
      "actions": [
        {
          "handler": "prometheus",
          "instances": [
            "tcpconnectionsclosed"
          ]
        }
      ]
    }
  },
  "istio-obj-97": {
    "apiVersion": "config.istio.io/v1alpha2",
    "kind": "handler",
    "metadata": {
      "name": "kubernetesenv",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-telemetry",
        "release": "istio"
      }
    },
    "spec": {
      "compiledAdapter": "kubernetesenv",
      "params": null
    }
  },
  "istio-obj-98": {
    "apiVersion": "config.istio.io/v1alpha2",
    "kind": "rule",
    "metadata": {
      "name": "kubeattrgenrulerule",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-telemetry",
        "release": "istio"
      }
    },
    "spec": {
      "actions": [
        {
          "handler": "kubernetesenv",
          "instances": [
            "attributes"
          ]
        }
      ]
    }
  },
  "istio-obj-99": {
    "apiVersion": "config.istio.io/v1alpha2",
    "kind": "rule",
    "metadata": {
      "name": "tcpkubeattrgenrulerule",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-telemetry",
        "release": "istio"
      }
    },
    "spec": {
      "match": "context.protocol == \"tcp\"",
      "actions": [
        {
          "handler": "kubernetesenv",
          "instances": [
            "attributes"
          ]
        }
      ]
    }
  },
  "istio-obj-100": {
    "apiVersion": "config.istio.io/v1alpha2",
    "kind": "instance",
    "metadata": {
      "name": "attributes",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-telemetry",
        "release": "istio"
      }
    },
    "spec": {
      "compiledTemplate": "kubernetes",
      "params": {
        "source_uid": "source.uid | \"\"",
        "source_ip": "source.ip | ip(\"0.0.0.0\")",
        "destination_uid": "destination.uid | \"\"",
        "destination_port": "destination.port | 0"
      },
      "attributeBindings": {
        "source.ip": "$out.source_pod_ip | ip(\"0.0.0.0\")",
        "source.uid": "$out.source_pod_uid | \"unknown\"",
        "source.labels": "$out.source_labels | emptyStringMap()",
        "source.name": "$out.source_pod_name | \"unknown\"",
        "source.namespace": "$out.source_namespace | \"default\"",
        "source.owner": "$out.source_owner | \"unknown\"",
        "source.serviceAccount": "$out.source_service_account_name | \"unknown\"",
        "source.workload.uid": "$out.source_workload_uid | \"unknown\"",
        "source.workload.name": "$out.source_workload_name | \"unknown\"",
        "source.workload.namespace": "$out.source_workload_namespace | \"unknown\"",
        "destination.ip": "$out.destination_pod_ip | ip(\"0.0.0.0\")",
        "destination.uid": "$out.destination_pod_uid | \"unknown\"",
        "destination.labels": "$out.destination_labels | emptyStringMap()",
        "destination.name": "$out.destination_pod_name | \"unknown\"",
        "destination.container.name": "$out.destination_container_name | \"unknown\"",
        "destination.namespace": "$out.destination_namespace | \"default\"",
        "destination.owner": "$out.destination_owner | \"unknown\"",
        "destination.serviceAccount": "$out.destination_service_account_name | \"unknown\"",
        "destination.workload.uid": "$out.destination_workload_uid | \"unknown\"",
        "destination.workload.name": "$out.destination_workload_name | \"unknown\"",
        "destination.workload.namespace": "$out.destination_workload_namespace | \"unknown\""
      }
    }
  },
  "istio-obj-101": {
    "apiVersion": "networking.istio.io/v1alpha3",
    "kind": "DestinationRule",
    "metadata": {
      "name": "istio-telemetry",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-telemetry",
        "release": "istio"
      }
    },
    "spec": {
      "host": "istio-telemetry.istio-system.svc.cluster.local",
      "trafficPolicy": {
        "portLevelSettings": [
          {
            "port": {
              "number": 15004
            },
            "tls": {
              "mode": "ISTIO_MUTUAL"
            }
          },
          {
            "port": {
              "number": 9091
            },
            "tls": {
              "mode": "DISABLE"
            }
          }
        ],
        "connectionPool": {
          "http": {
            "http2MaxRequests": 10000,
            "maxRequestsPerConnection": 10000
          }
        }
      }
    }
  },
  "istio-obj-102": {
    "apiVersion": "v1",
    "kind": "ConfigMap",
    "metadata": {
      "namespace": "istio-system",
      "name": "telemetry-envoy-config",
      "labels": {
        "release": "istio"
      }
    },
    "data": {
      "envoy.yaml.tmpl": "admin:\n  access_log_path: /dev/null\n  address:\n    socket_address:\n      address: 127.0.0.1\n      port_value: 15000\nstats_config:\n  use_all_default_tags: false\n  stats_tags:\n  - tag_name: cluster_name\n    regex: '^cluster\\.((.+?(\\..+?\\.svc\\.cluster\\.local)?)\\.)'\n  - tag_name: tcp_prefix\n    regex: '^tcp\\.((.*?)\\.)\\w+?$'\n  - tag_name: response_code\n    regex: '_rq(_(\\d{3}))$'\n  - tag_name: response_code_class\n    regex: '_rq(_(\\dxx))$'\n  - tag_name: http_conn_manager_listener_prefix\n    regex: '^listener(?=\\.).*?\\.http\\.(((?:[_.[:digit:]]*|[_\\[\\]aAbBcCdDeEfF[:digit:]]*))\\.)'\n  - tag_name: http_conn_manager_prefix\n    regex: '^http\\.(((?:[_.[:digit:]]*|[_\\[\\]aAbBcCdDeEfF[:digit:]]*))\\.)'\n  - tag_name: listener_address\n    regex: '^listener\\.(((?:[_.[:digit:]]*|[_\\[\\]aAbBcCdDeEfF[:digit:]]*))\\.)'\n\nstatic_resources:\n  clusters:\n  - name: prometheus_stats\n    type: STATIC\n    connect_timeout: 0.250s\n    lb_policy: ROUND_ROBIN\n    hosts:\n    - socket_address:\n        protocol: TCP\n        address: 127.0.0.1\n        port_value: 15000\n\n  - name: inbound_9092\n    circuit_breakers:\n      thresholds:\n      - max_connections: 100000\n        max_pending_requests: 100000\n        max_requests: 100000\n        max_retries: 3\n    connect_timeout: 1.000s\n    hosts:\n    - pipe:\n        path: /sock/mixer.socket\n    http2_protocol_options: {}\n\n  - name: out.galley.15019\n    http2_protocol_options: {}\n    connect_timeout: 1.000s\n    type: STRICT_DNS\n\n    circuit_breakers:\n      thresholds:\n        - max_connections: 100000\n          max_pending_requests: 100000\n          max_requests: 100000\n          max_retries: 3\n    hosts:\n      - socket_address:\n          address: istio-galley.istio-system\n          port_value: 15019\n    tls_context:\n      common_tls_context:\n        tls_certificates:\n        - certificate_chain:\n            filename: /etc/certs/cert-chain.pem\n          private_key:\n            filename: /etc/certs/key.pem\n        validation_context:\n          trusted_ca:\n            filename: /etc/certs/root-cert.pem\n          verify_subject_alt_name:\n          - spiffe://cluster.local/ns/istio-system/sa/istio-galley-service-account\n\n  listeners:\n  - name: \"15090\"\n    address:\n      socket_address:\n        protocol: TCP\n        address: 0.0.0.0\n        port_value: 15090\n    filter_chains:\n    - filters:\n      - name: envoy.http_connection_manager\n        config:\n          codec_type: AUTO\n          stat_prefix: stats\n          route_config:\n            virtual_hosts:\n            - name: backend\n              domains:\n              - '*'\n              routes:\n              - match:\n                  prefix: /stats/prometheus\n                route:\n                  cluster: prometheus_stats\n          http_filters:\n          - name: envoy.router\n\n  - name: \"15004\"\n    address:\n      socket_address:\n        address: 0.0.0.0\n        port_value: 15004\n    filter_chains:\n    - filters:\n      - config:\n          codec_type: HTTP2\n          http2_protocol_options:\n            max_concurrent_streams: 1073741824\n          generate_request_id: true\n          http_filters:\n          - config:\n              default_destination_service: istio-telemetry.istio-system.svc.cluster.local\n              service_configs:\n                istio-telemetry.istio-system.svc.cluster.local:\n                  disable_check_calls: true\n{{- if .DisableReportCalls }}\n                  disable_report_calls: true\n{{- end }}\n                  mixer_attributes:\n                    attributes:\n                      destination.service.host:\n                        string_value: istio-telemetry.istio-system.svc.cluster.local\n                      destination.service.uid:\n                        string_value: istio://istio-system/services/istio-telemetry\n                      destination.service.name:\n                        string_value: istio-telemetry\n                      destination.service.namespace:\n                        string_value: istio-system\n                      destination.uid:\n                        string_value: kubernetes://{{ .PodName }}.istio-system\n                      destination.namespace:\n                        string_value: istio-system\n                      destination.ip:\n                        bytes_value: {{ .PodIP }}\n                      destination.port:\n                        int64_value: 15004\n                      context.reporter.kind:\n                        string_value: inbound\n                      context.reporter.uid:\n                        string_value: kubernetes://{{ .PodName }}.istio-system\n              transport:\n                check_cluster: mixer_check_server\n                report_cluster: inbound_9092\n            name: mixer\n          - name: envoy.router\n          route_config:\n            name: \"15004\"\n            virtual_hosts:\n            - domains:\n              - '*'\n              name: istio-telemetry.istio-system.svc.cluster.local\n              routes:\n              - decorator:\n                  operation: Report\n                match:\n                  prefix: /\n                route:\n                  cluster: inbound_9092\n                  timeout: 0.000s\n          stat_prefix: \"15004\"\n        name: envoy.http_connection_manager\n      tls_context:\n        common_tls_context:\n          alpn_protocols:\n          - h2\n          tls_certificates:\n          - certificate_chain:\n              filename: /etc/certs/cert-chain.pem\n            private_key:\n              filename: /etc/certs/key.pem\n          validation_context:\n            trusted_ca:\n              filename: /etc/certs/root-cert.pem\n        require_client_certificate: true\n\n  - name: \"9091\"\n    address:\n      socket_address:\n        address: 0.0.0.0\n        port_value: 9091\n    filter_chains:\n    - filters:\n      - config:\n          codec_type: HTTP2\n          http2_protocol_options:\n            max_concurrent_streams: 1073741824\n          generate_request_id: true\n          http_filters:\n          - config:\n              default_destination_service: istio-telemetry.istio-system.svc.cluster.local\n              service_configs:\n                istio-telemetry.istio-system.svc.cluster.local:\n                  disable_check_calls: true\n{{- if .DisableReportCalls }}\n                  disable_report_calls: true\n{{- end }}\n                  mixer_attributes:\n                    attributes:\n                      destination.service.host:\n                        string_value: istio-telemetry.istio-system.svc.cluster.local\n                      destination.service.uid:\n                        string_value: istio://istio-system/services/istio-telemetry\n                      destination.service.name:\n                        string_value: istio-telemetry\n                      destination.service.namespace:\n                        string_value: istio-system\n                      destination.uid:\n                        string_value: kubernetes://{{ .PodName }}.istio-system\n                      destination.namespace:\n                        string_value: istio-system\n                      destination.ip:\n                        bytes_value: {{ .PodIP }}\n                      destination.port:\n                        int64_value: 9091\n                      context.reporter.kind:\n                        string_value: inbound\n                      context.reporter.uid:\n                        string_value: kubernetes://{{ .PodName }}.istio-system\n              transport:\n                check_cluster: mixer_check_server\n                report_cluster: inbound_9092\n            name: mixer\n          - name: envoy.router\n          route_config:\n            name: \"9091\"\n            virtual_hosts:\n            - domains:\n              - '*'\n              name: istio-telemetry.istio-system.svc.cluster.local\n              routes:\n              - decorator:\n                  operation: Report\n                match:\n                  prefix: /\n                route:\n                  cluster: inbound_9092\n                  timeout: 0.000s\n          stat_prefix: \"9091\"\n        name: envoy.http_connection_manager\n\n  - name: \"local.15019\"\n    address:\n      socket_address:\n        address: 127.0.0.1\n        port_value: 15019\n    filter_chains:\n      - filters:\n          - name: envoy.http_connection_manager\n            config:\n              codec_type: HTTP2\n              stat_prefix: \"15019\"\n              http2_protocol_options:\n                max_concurrent_streams: 1073741824\n\n              access_log:\n                - name: envoy.file_access_log\n                  config:\n                    path: /dev/stdout\n\n              http_filters:\n                - name: envoy.router\n\n              route_config:\n                name: \"15019\"\n\n                virtual_hosts:\n                  - name: istio-galley\n\n                    domains:\n                      - '*'\n\n                    routes:\n                      - match:\n                          prefix: /\n                        route:\n                          cluster: out.galley.15019\n                          timeout: 0.000s"
    }
  },
  "istio-obj-103": {
    "apiVersion": "apps/v1",
    "kind": "Deployment",
    "metadata": {
      "labels": {
        "app": "istio-mixer",
        "istio": "mixer",
        "release": "istio"
      },
      "name": "istio-telemetry",
      "namespace": "istio-system"
    },
    "spec": {
      "replicas": 1,
      "selector": {
        "matchLabels": {
          "istio": "mixer",
          "istio-mixer-type": "telemetry"
        }
      },
      "strategy": {
        "rollingUpdate": {
          "maxSurge": "100%",
          "maxUnavailable": "25%"
        }
      },
      "template": {
        "metadata": {
          "annotations": {
            "sidecar.istio.io/inject": "false"
          },
          "labels": {
            "app": "telemetry",
            "istio": "mixer",
            "istio-mixer-type": "telemetry"
          }
        },
        "spec": {
          "affinity": {
            "nodeAffinity": {
              "preferredDuringSchedulingIgnoredDuringExecution": [
                {
                  "preference": {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "amd64"
                        ]
                      }
                    ]
                  },
                  "weight": 2
                },
                {
                  "preference": {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "ppc64le"
                        ]
                      }
                    ]
                  },
                  "weight": 2
                },
                {
                  "preference": {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "s390x"
                        ]
                      }
                    ]
                  },
                  "weight": 2
                }
              ],
              "requiredDuringSchedulingIgnoredDuringExecution": {
                "nodeSelectorTerms": [
                  {
                    "matchExpressions": [
                      {
                        "key": "beta.kubernetes.io/arch",
                        "operator": "In",
                        "values": [
                          "amd64",
                          "ppc64le",
                          "s390x"
                        ]
                      }
                    ]
                  }
                ]
              }
            }
          },
          "containers": [
            {
              "args": [
                "--monitoringPort=15014",
                "--address",
                "unix:///sock/mixer.socket",
                "--log_output_level=default:info",
                "--configStoreURL=mcp://localhost:15019",
                "--configDefaultNamespace=istio-system",
                "--useAdapterCRDs=false",
                "--useTemplateCRDs=false",
                "--trace_zipkin_url=http://zipkin.istio-system:9411/api/v1/spans"
              ],
              "env": [
                {
                  "name": "POD_NAMESPACE",
                  "valueFrom": {
                    "fieldRef": {
                      "apiVersion": "v1",
                      "fieldPath": "metadata.namespace"
                    }
                  }
                },
                {
                  "name": "GOMAXPROCS",
                  "value": "6"
                }
              ],
              "image": "docker.io/istio/mixer:1.4.3",
              "imagePullPolicy": "IfNotPresent",
              "livenessProbe": {
                "httpGet": {
                  "path": "/version",
                  "port": 15014
                },
                "initialDelaySeconds": 5,
                "periodSeconds": 5
              },
              "name": "mixer",
              "ports": [
                {
                  "containerPort": 9091
                },
                {
                  "containerPort": 15014
                },
                {
                  "containerPort": 42422
                }
              ],
              "resources": {
                "limits": {
                  "cpu": "4800m",
                  "memory": "4G"
                },
                "requests": {
                  "cpu": "1000m",
                  "memory": "1G"
                }
              },
              "volumeMounts": [
                {
                  "mountPath": "/etc/certs",
                  "name": "istio-certs",
                  "readOnly": true
                },
                {
                  "mountPath": "/sock",
                  "name": "uds-socket"
                },
                {
                  "mountPath": "/var/run/secrets/istio.io/telemetry/adapter",
                  "name": "telemetry-adapter-secret",
                  "readOnly": true
                }
              ]
            },
            {
              "args": [
                "proxy",
                "--domain",
                "$(POD_NAMESPACE).svc.cluster.local",
                "--serviceCluster",
                "istio-telemetry",
                "--templateFile",
                "/var/lib/envoy/envoy.yaml.tmpl",
                "--controlPlaneAuthPolicy",
                "MUTUAL_TLS",
                "--trust-domain=cluster.local"
              ],
              "env": [
                {
                  "name": "POD_NAME",
                  "valueFrom": {
                    "fieldRef": {
                      "apiVersion": "v1",
                      "fieldPath": "metadata.name"
                    }
                  }
                },
                {
                  "name": "POD_NAMESPACE",
                  "valueFrom": {
                    "fieldRef": {
                      "apiVersion": "v1",
                      "fieldPath": "metadata.namespace"
                    }
                  }
                },
                {
                  "name": "INSTANCE_IP",
                  "valueFrom": {
                    "fieldRef": {
                      "apiVersion": "v1",
                      "fieldPath": "status.podIP"
                    }
                  }
                },
                {
                  "name": "SDS_ENABLED",
                  "value": "false"
                }
              ],
              "image": "docker.io/istio/proxyv2:1.4.3",
              "imagePullPolicy": "IfNotPresent",
              "name": "istio-proxy",
              "ports": [
                {
                  "containerPort": 15004
                },
                {
                  "containerPort": 15090,
                  "name": "http-envoy-prom",
                  "protocol": "TCP"
                }
              ],
              "resources": {
                "limits": {
                  "cpu": "2000m",
                  "memory": "1024Mi"
                },
                "requests": {
                  "cpu": "100m",
                  "memory": "128Mi"
                }
              },
              "volumeMounts": [
                {
                  "mountPath": "/var/lib/envoy",
                  "name": "telemetry-envoy-config"
                },
                {
                  "mountPath": "/etc/certs",
                  "name": "istio-certs",
                  "readOnly": true
                },
                {
                  "mountPath": "/sock",
                  "name": "uds-socket"
                }
              ]
            }
          ],
          "serviceAccountName": "istio-mixer-service-account",
          "volumes": [
            {
              "name": "istio-certs",
              "secret": {
                "optional": true,
                "secretName": "istio.istio-mixer-service-account"
              }
            },
            {
              "emptyDir": {},
              "name": "uds-socket"
            },
            {
              "name": "telemetry-adapter-secret",
              "secret": {
                "optional": true,
                "secretName": "telemetry-adapter-secret"
              }
            },
            {
              "configMap": {
                "name": "telemetry-envoy-config"
              },
              "name": "telemetry-envoy-config"
            }
          ]
        }
      }
    }
  },
  "istio-obj-104": {
    "apiVersion": "policy/v1beta1",
    "kind": "PodDisruptionBudget",
    "metadata": {
      "name": "istio-telemetry",
      "namespace": "istio-system",
      "labels": {
        "app": "telemetry",
        "release": "istio",
        "istio": "mixer",
        "istio-mixer-type": "telemetry"
      }
    },
    "spec": {
      "minAvailable": 1,
      "selector": {
        "matchLabels": {
          "app": "telemetry",
          "istio": "mixer",
          "istio-mixer-type": "telemetry"
        }
      }
    }
  },
  "istio-obj-105": {
    "apiVersion": "v1",
    "kind": "Service",
    "metadata": {
      "name": "istio-telemetry",
      "namespace": "istio-system",
      "labels": {
        "app": "mixer",
        "istio": "mixer",
        "release": "istio"
      }
    },
    "spec": {
      "ports": [
        {
          "name": "grpc-mixer",
          "port": 9091
        },
        {
          "name": "grpc-mixer-mtls",
          "port": 15004
        },
        {
          "name": "http-monitoring",
          "port": 15014
        },
        {
          "name": "prometheus",
          "port": 42422
        }
      ],
      "selector": {
        "istio": "mixer",
        "istio-mixer-type": "telemetry"
      }
    }
  },
  "istio-obj-106": {
    "apiVersion": "v1",
    "kind": "ServiceAccount",
    "metadata": {
      "name": "istio-mixer-service-account",
      "namespace": "istio-system",
      "labels": {
        "app": "istio-telemetry",
        "release": "istio"
      }
    }
  },
}
