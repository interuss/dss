variable "google_dns_managed_zone_name" {
  type        = string
  description = <<-EOT
  GCP DNS zone name to automatically manage DNS entries.

  Leave it empty to manage it manually.
  EOT
}