variable "google_delete_protection" {
  type        = bool
  default     = true
  description = <<-EOT
  Setting this to false make the GKE cluster deletable. Use with caution as this removes deletion protection.
  This setting should only be used to ease the management of test clusters.

  EOT
}
