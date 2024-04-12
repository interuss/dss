locals {
  workspace_folder = replace(replace(var.kubernetes_context_name, "/", "_"), ":", "_")
  # Replace ':' and '/' characters from folder name by underscores. Those characters are used by AWS for contexts.
  workspace_location = abspath("${path.module}/../../../../build/workspace/${local.workspace_folder}")
}



