locals {
  workspace_location = abspath("${path.module}/../../../build/workspace/${var.kubernetes.kubectl_cluster_context_name}")
}

resource "local_file" "tanka_config_main" {
  content = templatefile("${path.module}/templates/main.jsonnet.tmp", {
    root_path     = path.module
    VAR_NAMESPACE = var.dss_configuration.namespace
    VAR_CLUSTER_CONTEXT        = var.kubernetes.kubectl_cluster_context_name
    VAR_ENABLE_SCD             = false
    VAR_CRDB_HOSTNAME_SUFFIX   = var.dss_configuration.crdb_hostname_suffix
    VAR_CRDB_LOCALITY          = var.dss_configuration.crdb_locality
    VAR_CRDB_NODE_IPS          = join(",", [for i in var.kubernetes.crdb_nodes[*].ip : "'${i}'"])
    VAR_INGRESS_NAME           = var.kubernetes.ip_gateway
    VAR_CRDB_EXTERNAL_NODES    = join(",", [for a in var.dss_configuration.crdb_external_nodes : "'${a}'"])
    VAR_STORAGE_CLASS          = var.dss_configuration.storage_class
    VAR_DOCKER_IMAGE_NAME      = var.dss_configuration.image
    VAR_APP_HOSTNAME           = var.dss_configuration.app_hostname
    VAR_PUBLIC_KEY_PEM_PATH    = var.dss_configuration.public_key_pem_path
    VAR_JWKS_ENDPOINT          = var.dss_configuration.jwks_endpoint
    VAR_JWKS_KEY_ID            = var.dss_configuration.jwks_key_id
    VAR_DESIRED_RID_DB_VERSION = "4.0.0"
    VAR_DESIRED_SCD_DB_VERSION = "3.1.0"
    VAR_SHOULD_INIT            = var.dss_configuration.should_init
  })
  filename = "${local.workspace_location}/main.jsonnet"
}

resource "local_file" "tanka_config_spec" {
  content = templatefile("${path.module}/templates/spec.json.tmp", {
    root_path       = path.module
    VAR_NAMESPACE   = var.dss_configuration.namespace
    cluster_context = var.kubernetes.kubectl_cluster_context_name
    api_server      = var.kubernetes.api_endpoint
  })
  filename = "${local.workspace_location}/spec.json"
}

resource "local_file" "make_certs" {
  content = templatefile("${path.module}/templates/make-certs.sh.tmp", {
    cluster_context = var.kubernetes.kubectl_cluster_context_name
    namespace       = var.dss_configuration.namespace
    node_address    = join(" ", var.kubernetes.node_addresses)
  })
  filename = "${local.workspace_location}/make-certs.sh"
}

resource "local_file" "apply_certs" {
  content = templatefile("${path.module}/templates/apply-certs.sh.tmp", {
    cluster_context = var.kubernetes.kubectl_cluster_context_name
    namespace       = var.dss_configuration.namespace
  })
  filename = "${local.workspace_location}/apply-certs.sh"
}

resource "local_file" "get_credentials" {
  content = templatefile("${path.module}/templates/get-credentials.sh.tmp", {
    get_credentials_cmd = var.kubernetes.get_credentials_cmd
  })
  filename = "${local.workspace_location}/get-credentials.sh"
}