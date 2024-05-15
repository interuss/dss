variable "node_count" {
  type        = number
  description = <<-EOT
    Number of Kubernetes nodes which should correspond to the desired CockroachDB nodes.
    Currently, only single node or three nodes deployments are supported.

    Example: `3`
  EOT

  validation {
    condition     = contains([1, 3], var.node_count)
    error_message = "Currently, only 1 node or 3 nodes deployments are supported."
  }
}
