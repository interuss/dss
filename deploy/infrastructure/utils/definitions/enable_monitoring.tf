variable "enable_monitoring" {
  type        = bool
  default     = false
  description = <<-EOT
  Set to true to enable monitoring stack with prometheus / grafana.

  Example: `true`
  EOT
}
