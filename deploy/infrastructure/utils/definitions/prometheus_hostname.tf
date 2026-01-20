variable "prometheus_hostname" {
  type        = string
  default     = ""
  description = <<-EOT
  Domain used to expose prometheus on an external endpoint.

  Leave empty to disable exposition of prometheus publicly.

  Example: `prometheus.dss.example.com`

  EOT
}
