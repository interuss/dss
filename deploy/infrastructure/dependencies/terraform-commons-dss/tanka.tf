locals {
  tanka_workspace_location = abspath("${path.module}/../../../../deploy/services/tanka/workspace/${local.workspace_folder}")
}

resource "local_file" "tanka_config_main" {
  content = templatefile("${path.module}/templates/main.jsonnet.tmp", {
    root_path                         = path.module
    VAR_NAMESPACE                     = var.kubernetes_namespace
    VAR_CLUSTER_CONTEXT               = var.kubernetes_context_name
    VAR_ENABLE_SCD                    = var.enable_scd
    VAR_DB_HOSTNAME_SUFFIX            = var.db_hostname_suffix
    VAR_LOCALITY                      = var.locality
    VAR_DATASTORE                     = var.datastore_type
    VAR_CRDB_NODE_IPS                 = join(",", [for i in var.crdb_internal_nodes[*].ip : "'${i}'"])
    VAR_INGRESS_NAME                  = var.ip_gateway
    VAR_CRDB_EXTERNAL_NODES           = join(",", [for a in var.crdb_external_nodes : "'${a}'"])
    VAR_STORAGE_CLASS                 = var.kubernetes_storage_class
    VAR_DOCKER_IMAGE_NAME             = var.image
    VAR_CRDB_DOCKER_IMAGE_NAME        = "cockroachdb/cockroach:${var.crdb_image_tag}"
    VAR_YUGABYTE_CLOUD                = var.yugabyte_cloud
    VAR_YUGABYTE_REGION               = var.yugabyte_region
    VAR_YUGABYTE_ZONE                 = var.yugabyte_zone
    VAR_YUGABYTE_MASTER_NODE_IPS      = join(",", [for a in var.yugabyte_internal_masters_nodes[*].ip : "'${a}'"])
    VAR_YUGABYTE_TSERVER_NODE_IPS     = join(",", [for a in var.yugabyte_internal_tservers_nodes[*].ip : "'${a}'"])
    VAR_YUGABYTE_MASTER_ADDRESSES     = join(",", [for m in concat([for i in range(var.node_count) : format("%s.master.${var.db_hostname_suffix}", i)], var.yugabyte_external_nodes) : "'${m}'"])
    VAR_YUGABYTE_MASTER_BIND_ADDRESS  = "$${HOSTNAMENO}.master.${var.db_hostname_suffix}"
    VAR_YUGABYTE_TSERVER_BIND_ADDRESS = "$${HOSTNAMENO}.tserver.${var.db_hostname_suffix}"
    VAR_YUGABYTE_DOCKER_IMAGE_NAME    = "yugabytedb/yugabyte:2.25.1.0-b381" // TODO: This should be an option
    VAR_YUGABYTE_LIGHT_RESOURCES      = var.yugabyte_light_resources
    VAR_APP_HOSTNAME                  = var.app_hostname
    VAR_PUBLIC_KEY_PEM_PATH           = var.authorization.public_key_pem_path != null ? var.authorization.public_key_pem_path : ""
    VAR_JWKS_ENDPOINT                 = var.authorization.jwks != null ? var.authorization.jwks.endpoint : ""
    VAR_JWKS_KEY_ID                   = var.authorization.jwks != null ? var.authorization.jwks.key_id : ""
    VAR_DESIRED_RID_DB_VERSION        = local.rid_db_schema
    VAR_DESIRED_SCD_DB_VERSION        = local.scd_db_schema
    VAR_DESIRED_AUX_DB_VERSION        = local.aux_db_schema
    VAR_SHOULD_INIT                   = var.should_init
    VAR_DOCKER_IMAGE_PULL_SECRET      = var.image_pull_secret != null ? var.image_pull_secret : ""
    VAR_CLOUD_PROVIDER                = var.kubernetes_cloud_provider_name
    VAR_CERT_NAME                     = var.gateway_cert_name
    VAR_SUBNET                        = var.workload_subnet
    VAR_SSL_POLICY                    = var.ssl_policy
  })
  filename = "${local.tanka_workspace_location}/main.jsonnet"
}

resource "local_file" "tanka_config_spec" {
  content = templatefile("${path.module}/templates/spec.json.tmp", {
    root_path       = path.module
    namespace       = var.kubernetes_namespace
    cluster_context = var.kubernetes_context_name
    api_server      = var.kubernetes_api_endpoint
  })
  filename = "${local.tanka_workspace_location}/spec.json"
}
