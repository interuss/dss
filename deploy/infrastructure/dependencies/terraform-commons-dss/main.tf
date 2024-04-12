locals {
  workspace_folder = replace(replace(var.kubernetes_context_name, "/", "_"), ":", "_")
  # Replace ':' and '/' characters from folder name by underscores. Those characters are used by AWS for contexts.
  workspace_location = abspath("${path.module}/../../../../build/workspace/${local.workspace_folder}")

  # Tanka defines itself the variables below. For helm, since we are using the official helm CRDB chart,
  # the following has to be provided and constructed here.
  helm_crdb_statefulset_name = "dss-cockroachdb"
  helm_nodes_to_join         = concat(["${local.helm_crdb_statefulset_name}-0.${local.helm_crdb_statefulset_name}"], var.crdb_external_nodes)
}



