variable "aws_kubernetes_storage_class" {
  type        = string
  description = <<-EOT
  AWS Elastic Kubernetes Service Storage Class to use for datastores and Prometheus persistent volumes.
  See https://docs.aws.amazon.com/eks/latest/userguide/storage-classes.html for more details and
  available options.

  Depending on your use case, performance may be significantly improved with higher-tier storage classes, though this should be balanced against the associated costs.

  Both CockroachDB and YugabyteDB recommend at least `gp3` for production workloads. Use `gp2` for testing only, or consider `io2` for high-throughput scenarios.

  See https://www.cockroachlabs.com/docs/v24.1/recommended-production-settings#aws and https://docs.yugabyte.com/stable/deploy/checklist/#amazon-web-services-aws for database-specific recommendations.

  Example: `gp3` for production and `gp2` for development.
  EOT
}
