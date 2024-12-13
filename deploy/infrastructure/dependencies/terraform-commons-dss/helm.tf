locals {
  # Tanka defines itself the variable below. For helm, since we are using the official helm CRDB chart,
  # the following variable has to be provided here.
  helm_crdb_statefulset_name = "dss-cockroachdb"
  locality_args = var.crdb_region != "" ? "region=${var.crdb_region},zone=${var.crdb_locality}" : "zone=${var.crdb_locality}"
}

resource "local_file" "helm_chart_values" {
  filename = "${local.workspace_location}/helm_values.yml"
  content = yamlencode({
    cockroachdb = {
      image = {
        tag = var.crdb_image_tag
      }
      fullnameOverride = local.helm_crdb_statefulset_name

      conf = {
        join         = var.crdb_external_nodes
        cluster-name = var.crdb_cluster_name
        single-node  = false # Always false. Even with 1 replica, we would expect to keep the ability to pool it with another cluster.
        locality     = locals.locality_args
      }

      statefulset = {
        replicas = length(var.crdb_internal_nodes)
        args = [
          "--locality-advertise-addr=${locals.locality_args}@$(hostname -f)",
          "--advertise-addr=$${HOSTNAME##*-}.${var.crdb_hostname_suffix}"
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
  })
}
