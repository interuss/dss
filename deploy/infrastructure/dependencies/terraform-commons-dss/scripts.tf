
resource "local_file" "make_certs" {
  content = templatefile("${path.module}/templates/make-certs.sh.tmp", {
    cluster_context = var.kubernetes_context_name
    namespace       = var.kubernetes_namespace
    node_address    = join(" ", var.crdb_internal_nodes[*].dns)
    joining_pool    = length(var.crdb_external_nodes) > 0
  })
  filename = "${local.workspace_location}/make-certs.sh"
}

resource "local_file" "apply_certs" {
  content = templatefile("${path.module}/templates/apply-certs.sh.tmp", {
    cluster_context = var.kubernetes_context_name
    namespace       = var.kubernetes_namespace
  })
  filename = "${local.workspace_location}/apply-certs.sh"
}

resource "local_file" "get_credentials" {
  content = templatefile("${path.module}/templates/get-credentials.sh.tmp", {
    get_credentials_cmd = var.kubernetes_get_credentials_cmd
  })
  filename = "${local.workspace_location}/get-credentials.sh"
}
