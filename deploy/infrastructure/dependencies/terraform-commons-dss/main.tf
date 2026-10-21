locals {
  workspace_folder = replace(replace(var.kubernetes_context_name, "/", "_"), ":", "_")
  # Replace ':' and '/' characters from folder name by underscores. Those characters are used by AWS for contexts.
  workspace_location = abspath("${path.module}/../../../../build/workspace/${local.workspace_folder}")

  evict_minute       = parseint(substr(sha256(var.locality), 0, 8), 16) % 60
  evict_scd_schedule = var.evict_scd_schedule != "" ? var.evict_scd_schedule : "${local.evict_minute} 2 * * *"
  evict_rid_schedule = var.evict_rid_schedule != "" ? var.evict_rid_schedule : "${local.evict_minute} * * * *"
}
