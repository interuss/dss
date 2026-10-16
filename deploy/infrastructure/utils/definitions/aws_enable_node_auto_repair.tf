variable "aws_enable_node_auto_repair" {
  type        = bool
  description = <<-EOT
  Enable EKS node auto repair on the managed node group.

  When enabled, EKS monitors the health of the worker nodes and automatically replaces nodes
  which become unhealthy (for instance, nodes stuck in `NotReady`).
  This is particularly important for datastores, since a datastore pod cannot be rescheduled on
  another node already hosting a datastore pod.

  See https://docs.aws.amazon.com/eks/latest/userguide/node-health.html for more details.

  Example: `true`
  EOT

  default = true
}
