variable "crdb_memory_request" {
  type        = string
  description = <<-EOT
  Kubernetes memory request of the CockroachDB containers.
  Not set by default.

  Example: `14Gi`
  EOT

  default = ""
}
