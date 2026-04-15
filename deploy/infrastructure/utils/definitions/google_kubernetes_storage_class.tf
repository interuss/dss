variable "google_kubernetes_storage_class" {
  type        = string
  description = <<-EOT
  GCP Kubernetes Storage Class to use for datastores and Prometheus persistent volumes.
  See https://docs.cloud.google.com/compute/docs/disks/persistent-disks#disk-types for more details and
  available options.

  Depending on your use case, performance may be significantly improved with higher-tier storage classes, though this should be balanced against the associated costs.

  Both CockroachDB and YugabyteDB recommend SSDs for production workloads, configured via the `premium-rwo` storage class. Use `standard` for testing only.

  See https://www.cockroachlabs.com/docs/v24.1/recommended-production-settings#gcp and https://docs.yugabyte.com/stable/deploy/checklist/#google-cloud for database-specific recommendations.

  Example: `premium-rwo` for production and `standard` for development.
  EOT
}
