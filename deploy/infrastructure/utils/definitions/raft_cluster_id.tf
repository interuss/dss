variable "raft_cluster_id" {
  type        = number
  description = <<-EOT
  Raft cluster ID. Must be identical on all nodes of the pool.

  Only used with the raft datastore.

  Example: `1`
  EOT

  validation {
    condition     = var.raft_cluster_id >= 1
    error_message = "raft_cluster_id must be equal to or greater than 1."
  }

  default = 1
}
