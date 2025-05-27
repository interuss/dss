locals {
  # Tanka defines itself the variable below. For helm, since we are using the official helm CRDB chart,
  # the following variable has to be provided here.
  helm_crdb_statefulset_name = "dss-cockroachdb"
}

resource "local_file" "helm_chart_values" {
  filename = "${local.workspace_location}/helm_values.yml"
  content = var.datastore_type == "cockroachdb" ? yamlencode({
    cockroachdb = {
      image = {
        tag = var.crdb_image_tag
      }
      fullnameOverride = local.helm_crdb_statefulset_name

      conf = {
        join         = var.crdb_external_nodes
        cluster-name = var.crdb_cluster_name
        single-node  = false # Always false. Even with 1 replica, we would expect to keep the ability to pool it with another cluster.
        locality     = "zone=${var.crdb_locality}"
      }

      statefulset = {
        replicas = length(var.crdb_internal_nodes)
        args = [
          "--locality-advertise-addr=zone=${var.crdb_locality}@$(hostname -f)",
          "--advertise-addr=$${HOSTNAME##*-}.${var.db_hostname_suffix}"
        ]
      }

      storage = {
        persistentVolume = {
          storageClass = var.kubernetes_storage_class
        }
      }
    }

    loadBalancers = {
      cockroachdbNodes = [
        for ip in var.crdb_internal_nodes[*].ip :
        {
          ip     = ip
          subnet = var.workload_subnet
        }
      ]

      dssGateway = {
        ip        = var.ip_gateway
        subnet    = var.workload_subnet
        certName  = var.gateway_cert_name
        sslPolicy = var.ssl_policy
      }
    }

    dss = {
      image = var.image

      conf = {
        pubKeys = [
          "/test-certs/auth2.pem"
        ]
        jwksEndpoint = var.authorization.jwks != null ? var.authorization.jwks.endpoint : ""
        jwksKeyIds   = var.authorization.jwks != null ? [var.authorization.jwks.key_id] : []
        hostname     = var.app_hostname
        enableScd    = var.enable_scd
      }
    }

    global = {
      cloudProvider = var.kubernetes_cloud_provider_name
    }
}) : yamlencode({
    cockroachdb = {
      enabled = false
      image = {
        tag = "dummy"
      }
      fullnameOverride = "dummy"
      conf = {
        cluster-name = "dummy"
        locality = "dummy"
      }
      statefulset = {}
    }
    yugabyte = {
      enabled = true
      Image = {
        tag = "2.25.2.0-b359"
      }
      nameOverride = "dss-yugabyte"

      resource = var.yugabyte_light_resources ? {
        master = {
          requests = {
            cpu = "0.1"
            memory = "0.5G"
          }
        }
        tserver = {
          requests = {
            cpu = "0.1"
            memory = "0.5G"
          }
        }
      } : {}
      enableLoadBalancer = false

      storage = {
        master = {
          storageClass = var.kubernetes_storage_class
        }
        tserver = {
          storageClass = var.kubernetes_storage_class
        }
      }

      preflight = {
        skipUlimit = true
      }

      master = {
        extraEnv = [{
          name = "HOSTNAMENO"
          valueFrom =  {
            fieldRef = {
              fieldPath = "metadata.labels['apps.kubernetes.io/pod-index']"
            }
          }
        }]
        serverBroadcastAddress: "$${HOSTNAMENO}.master.${var.db_hostname_suffix}"
        rpcBindAddress: "$${HOSTNAMENO}.master.${var.db_hostname_suffix}"
        preCommands: "sed -E \"/\\.svc\\.cluster\\.local/ s/^([0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+)([[:space:]]+)/\\1 $(echo \"$${HOSTNAMENO}.master.${var.db_hostname_suffix}\" | sed 's/[\\/&]/\\\\&/g')\\2/\" /etc/hosts > /tmp/newhosts && /bin/cp /tmp/newhosts /etc/hosts && \\"
      }

      tserver = {
        extraEnv = [{
          name = "HOSTNAMENO"
          valueFrom =  {
            fieldRef = {
              fieldPath = "metadata.labels['apps.kubernetes.io/pod-index']"
            }
          }
        }]
        serverBroadcastAddress: "$${HOSTNAMENO}.tserver.${var.db_hostname_suffix}"
        rpcBindAddress: "$${HOSTNAMENO}.tserver.${var.db_hostname_suffix}"
        preCommands: "sed -E \"/\\.svc\\.cluster\\.local/ s/^([0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+)([[:space:]]+)/\\1 $(echo \"$${HOSTNAMENO}.tserver.${var.db_hostname_suffix}\" | sed 's/[\\/&]/\\\\&/g')\\2/\" /etc/hosts > /tmp/newhosts && /bin/cp /tmp/newhosts /etc/hosts && \\"
      }

      gflags = {
        master = {
          placement_cloud: var.yugabyte_cloud
          placement_region: var.yugabyte_region
          placement_zone: var.yugabyte_zone
          use_private_ip: "zone"
        }
        tserver = {
          placement_cloud: var.yugabyte_cloud
          placement_region: var.yugabyte_region
          placement_zone: var.yugabyte_zone
          use_private_ip: "zone"
        }
      }

      isMultiAz = true
      masterAddresses = join(",", ["0.master.${var.db_hostname_suffix},1.master.${var.db_hostname_suffix},2.master.${var.db_hostname_suffix}", join(",", var.yugabyte_external_nodes)])
    }

    loadBalancers = {
      cockroachdbNodes = []

      yugabyteMasterNodes = [
        for ip in var.yugabyte_internal_masters_nodes[*].ip :
        {
          ip     = ip
          subnet = var.workload_subnet
        }
      ]

      yugabyteTserverNodes = [
        for ip in var.yugabyte_internal_tservers_nodes[*].ip :
        {
          ip     = ip
          subnet = var.workload_subnet
        }
      ]

      dssGateway = {
        ip        = var.ip_gateway
        subnet    = var.workload_subnet
        certName  = var.gateway_cert_name
        sslPolicy = var.ssl_policy
      }
    }

    dss = {
      image = var.image

      conf = {
        pubKeys = [
          "/test-certs/auth2.pem"
        ]
        jwksEndpoint = var.authorization.jwks != null ? var.authorization.jwks.endpoint : ""
        jwksKeyIds   = var.authorization.jwks != null ? [var.authorization.jwks.key_id] : []
        hostname     = var.app_hostname
        enableScd    = var.enable_scd
      }
    }

    global = {
      cloudProvider = var.kubernetes_cloud_provider_name
    }
})

}
