variable "kubernetes_storage_class" {
  type        = string
  description = <<-EOT
  Kubernetes Storage Class to use for CockroachDB and Prometheus volumes. You can
  check your cluster's possible values with kubectl get storageclass.
  This value may be cloud provider specific.
  Example for GCP: standard
  EOT
}