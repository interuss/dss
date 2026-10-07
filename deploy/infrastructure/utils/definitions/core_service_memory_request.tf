variable "core_service_memory_request" {
  type        = string
  description = <<-EOT
  Kubernetes memory request of the core-service containers.
  Not set by default.

  Example: `2Gi`
  EOT

  default = ""
}
