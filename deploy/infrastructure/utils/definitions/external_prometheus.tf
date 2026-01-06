variable "external_prometheus" {
  type        = string
  default     = ""
  description = <<-EOT
  Domain used to expose prometheus on an external endpoint.

  Leave empty to disable exposition of prometheus publicly.

  Only supported in helm deployments.

  Example: `prometheus.dss.example.com`

  EOT
}
