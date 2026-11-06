variable "datastore_type" {
  type        = string
  description = <<-EOT
  Type of datastore used

  Supported technologies: cockroachdb, yugabyte, raft
  EOT

  validation {
    condition     = contains(["cockroachdb", "yugabyte", "raft"], var.datastore_type)
    error_message = "Supported technologies: cockroachdb, yugabyte, raft"
  }

  default = "cockroachdb"
}
