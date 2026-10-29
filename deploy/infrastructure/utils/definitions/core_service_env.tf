variable "core_service_env" {
  type        = map(string)
  description = <<-EOT
  Additional environment variables of the core-service containers, for instance to tune the Go runtime.

  Example: `{ GOGC = "80", GOMEMLIMIT = "1800MiB" }`
  EOT

  default = {}
}
