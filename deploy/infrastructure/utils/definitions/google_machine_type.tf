variable "google_machine_type" {
  type        = string
  description = <<-EOT
    GCP machine type used for the Kubernetes node pool.
    Example: `n2-standard-4` for production, `e2-medium` for development
  EOT
}