variable "core_service_cpu_request" {
  type        = string
  description = <<-EOT
  Kubernetes CPU request of the core-service containers.
  Not set by default.

  Example: `0.5`
  EOT

  default = ""
}
