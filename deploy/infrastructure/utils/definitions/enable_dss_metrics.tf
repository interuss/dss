variable "enable_dss_metrics" {
  type        = bool
  default     = false
  description = <<-EOT
  Enable DSS's prometheus metric.

  Require DSS version to be at least 0.23.0.

  Example: `true`
  EOT
}
