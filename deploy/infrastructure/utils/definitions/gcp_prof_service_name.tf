variable "gcp_prof_service_name" {
  type        = string
  default     = ""
  description = <<-EOT
  Service name for the Go profiler of the core-service. The profiler is disabled when empty.

  Example: `dss-core-service`
  EOT
}
