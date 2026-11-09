variable "raft_node_ids" {
  type        = list(number)
  description = <<-EOT
  Raft node IDs of the nodes deployed by this instance, in pod-ordinal order.
  Must contain exactly `node_count` entries. IDs must be unique within the pool and never reused:
  do not change or reorder them once the pool has been bootstrapped.

  Only used with the raft datastore.

  Example: `[1, 2, 3]`
  EOT

  validation {
    condition     = var.datastore_type != "raft" || length(var.raft_node_ids) == var.node_count
    error_message = "raft_node_ids must contain exactly node_count entries."
  }

  default = []
}
