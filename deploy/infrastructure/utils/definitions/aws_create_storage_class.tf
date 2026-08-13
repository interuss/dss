variable "aws_create_storage_class" {
  type        = bool
  description = <<-EOT
  Create storage class in cluster.

  For now, create a gp3 storage class in the cluster and set it as the default one.

  Example: `true`
  EOT

  default = true
}
