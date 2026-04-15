variable "google_machine_type" {
  type        = string
  description = <<-EOT
  GCP machine type used for the Kubernetes node pool.
  See https://docs.cloud.google.com/compute/docs/machine-resource for more details and available options.

  Depending on your use case, performance may be significantly improved with higher-tier instances, though this should be balanced against the associated costs.

  Both CockroachDB and YugabyteDB recommend `n2` instances for production. Use `f1` and `g1` instances for testing only.

  See https://www.cockroachlabs.com/docs/v24.1/recommended-production-settings#gcp and https://docs.yugabyte.com/stable/deploy/checklist/#google-cloud for database-specific recommendations.

  Example: `n2-standard-4` for production, `e2-medium` for development.

  EOT
}
