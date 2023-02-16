variable "aws_kubernetes_storage_class" {
  type        = string
  description = <<-EOT
  AWS Elastic Kubernetes Service Storage Class to use for CockroachDB and Prometheus persistent volumes.
  See https://docs.aws.amazon.com/eks/latest/userguide/storage-classes.html for more details and
  available options.

  Example: `gp2`
  EOT
}