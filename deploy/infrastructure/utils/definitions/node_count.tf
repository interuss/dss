variable "node_count" {
  type        = number
  description = <<-EOT
  Number of Kubernetes nodes which should correspond to the desired CockroachDB nodes.
  Currently, only single node or three nodes deployments are supported.

  Example: `3`
  EOT

  validation {
    condition     = (var.datastore_type == "cockroachdb" && contains([1, 3], var.node_count)) || (var.datastore_type == "yugabyte" && var.node_count > 0)
    error_message = "Currently, only 1 node or 3 nodes deployments are supported for CockroachDB. If you use Yugabyte, you need to have at least one node."
  }
}
