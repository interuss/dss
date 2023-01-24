variable "google_kubernetes_storage_class" {
  type        = string
  description = <<-EOT
  GCP Kubernetes Storage Class to use for CockroachDB and Prometheus persistent volumes.
  See https://cloud.google.com/kubernetes-engine/docs/concepts/persistent-volumes for more details and
  available options.

  Example: `standard`
  EOT
}