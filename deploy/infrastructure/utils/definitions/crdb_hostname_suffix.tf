variable "crdb_hostname_suffix" {
  type        = string
  default     = ""
  description = <<-EOT
  This variable has been renamed to db_hostname_suffix and is left to warn users about migration.

  EOT

  validation {
    condition     = var.crdb_hostname_suffix == ""
    error_message = "crdb_hostname_suffix value is not supported anymore. Use `db_hostname_suffix` for similar behavior."
  }
}
