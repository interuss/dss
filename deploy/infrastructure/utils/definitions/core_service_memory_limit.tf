variable "core_service_memory_limit" {
  type        = string
  description = <<-EOT
  Kubernetes memory limit of the core-service containers.
  Not set by default.

  Example: `2Gi`
  EOT

  default = ""
}
