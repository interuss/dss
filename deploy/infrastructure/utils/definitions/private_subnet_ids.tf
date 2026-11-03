variable "private_subnet_ids" {
  type        = list(string)
  default     = []
  description = <<-EOT
  IDs of the private subnets to use for the Kubernetes cluster, instead of creating new ones.
  For backward compatibility, the default is an empty list.

  Example: `["subnet-01234567890123456", "subnet-01234567890123456"]`
  EOT
}
