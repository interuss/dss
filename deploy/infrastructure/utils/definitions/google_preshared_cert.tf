variable "google_preshared_cert" {
  type        = string
  default     = ""
  description = <<-EOT
  Name of a pre-shared SSL certificate of the Google Cloud project, used by the DSS gateway
  instead of a Google-managed certificate. A Google-managed certificate is used when empty.

  Example: `dss-certificate`
  EOT
}
