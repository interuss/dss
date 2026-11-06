locals {
  workspace_folder = replace(replace(var.kubernetes_context_name, "/", "_"), ":", "_")
  # Replace ':' and '/' characters from folder name by underscores. Those characters are used by AWS for contexts.
  workspace_location = abspath("${path.module}/../../../../build/workspace/${local.workspace_folder}")

  evict_minute       = parseint(substr(sha256(var.locality), 0, 8), 16) % 60
  evict_scd_schedule = var.evict_scd_schedule != "" ? var.evict_scd_schedule : "${local.evict_minute} 2 * * *"
  evict_rid_schedule = var.evict_rid_schedule != "" ? var.evict_rid_schedule : "${local.evict_minute} * * * *"
}

locals {
  # Kubernetes container resources of CockroachDB, without the requests and limits left unset.
  # Only used by Helm, since the official CockroachDB chart expects a resources map.
  helm_crdb_resources = { for k, v in {
    requests = { for k, v in { cpu = var.crdb_cpu_request, memory = var.crdb_memory_request } : k => v if v != "" }
    limits   = { for k, v in { cpu = var.crdb_cpu_limit, memory = var.crdb_memory_limit } : k => v if v != "" }
  } : k => v if length(v) > 0 }
}
