variable "core_service_cpu_limit" {
  type        = string
  description = <<-EOT
  Kubernetes CPU limit of the core-service containers.
  Not set by default.

  Example: `1`
  EOT

  default = ""
}
