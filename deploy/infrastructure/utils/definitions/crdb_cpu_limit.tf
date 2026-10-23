variable "crdb_cpu_limit" {
  type        = string
  description = <<-EOT
  Kubernetes CPU limit of the CockroachDB containers.
  Not set by default.

  Example: `3.5`
  EOT

  default = ""
}
