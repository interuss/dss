variable "crdb_cpu_request" {
  type        = string
  description = <<-EOT
  Kubernetes CPU request of the CockroachDB containers.
  Not set by default.

  Example: `3.5`
  EOT

  default = ""
}
