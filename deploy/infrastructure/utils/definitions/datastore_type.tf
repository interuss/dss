variable "datastore_type" {
  type        = string
  description = <<-EOT
  Type of datastore used

  Supported technologies: cockroachdb, yugabyte
  EOT

  validation {
    condition     = contains(["cockroachdb", "yugabyte"], var.datastore_type)
    error_message = "Supported technologies: cockroachdb, yugabyte"
  }

  default = "cockroachdb"
}
