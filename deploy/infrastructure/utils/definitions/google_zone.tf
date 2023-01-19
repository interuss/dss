variable "google_zone" {
  type = string
  description = <<-EOT
    GCP zone hosting the kubernetes cluster
    List of available zones: https://cloud.google.com/compute/docs/regions-zones#available

    Example: `europe-west6-a`
  EOT
}