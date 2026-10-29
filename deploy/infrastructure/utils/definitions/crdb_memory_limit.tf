variable "crdb_memory_limit" {
  type        = string
  description = <<-EOT
  Kubernetes memory limit of the CockroachDB containers.
  Not set by default.

  Example: `14Gi`
  EOT

  default = ""
}
