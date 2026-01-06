variable "external_prometheus_allowed_ips" {
  type        = list(string)
  default     = []
  description = <<-EOT
  List of subnets allowed to connect to the external prometheus.

  Only supported in helm deployments.

  Example: `1.2.3.4/24,42.42.42.42/32`

  EOT
}
