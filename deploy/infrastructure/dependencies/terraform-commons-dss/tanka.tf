resource "local_file" "tanka_config_main" {
  content = templatefile("${path.module}/templates/main.jsonnet.tmp", {
    root_path                    = path.module
    VAR_NAMESPACE                = var.kubernetes_namespace
    VAR_CLUSTER_CONTEXT          = var.kubernetes_context_name
    VAR_ENABLE_SCD               = var.enable_scd
    VAR_CRDB_HOSTNAME_SUFFIX     = var.crdb_hostname_suffix
    VAR_CRDB_LOCALITY            = var.crdb_locality
    VAR_CRDB_NODE_IPS            = join(",", [for i in var.crdb_internal_nodes[*].ip : "'${i}'"])
    VAR_INGRESS_NAME             = var.ip_gateway
    VAR_CRDB_EXTERNAL_NODES      = join(",", [for a in var.crdb_external_nodes : "'${a}'"])
    VAR_STORAGE_CLASS            = var.kubernetes_storage_class
    VAR_DOCKER_IMAGE_NAME        = local.image
    VAR_APP_HOSTNAME             = var.app_hostname
    VAR_PUBLIC_KEY_PEM_PATH      = var.authorization.public_key_pem_path != null ? var.authorization.public_key_pem_path : ""
    VAR_JWKS_ENDPOINT            = var.authorization.jwks != null ? var.authorization.jwks.endpoint : ""
    VAR_JWKS_KEY_ID              = var.authorization.jwks != null ? var.authorization.jwks.key_id : ""
    VAR_DESIRED_RID_DB_VERSION   = local.rid_db_schema
    VAR_DESIRED_SCD_DB_VERSION   = local.scd_db_schema
    VAR_SHOULD_INIT              = var.should_init
    VAR_DOCKER_IMAGE_PULL_SECRET = var.image_pull_secret != null ? var.image_pull_secret : ""
    VAR_CLOUD_PROVIDER           = var.kubernetes_cloud_provider_name
    VAR_CERT_NAME                = var.gateway_cert_name
    VAR_SUBNET                   = var.workload_subnet
    VAR_SSL_POLICY               = var.ssl_policy
  })
  filename = "${local.workspace_location}/main.jsonnet"
}

resource "local_file" "tanka_config_spec" {
  content = templatefile("${path.module}/templates/spec.json.tmp", {
    root_path       = path.module
    namespace       = var.kubernetes_namespace
    cluster_context = var.kubernetes_context_name
    api_server      = var.kubernetes_api_endpoint
  })
  filename = "${local.workspace_location}/spec.json"
}
