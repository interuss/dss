variable "crdb_locality" {
  type        = string
  default     = ""
  description = <<-EOT
  This variable has been renamed to locality and is left to warn users about migration.

  EOT

  validation {
    condition     = var.crdb_locality == ""
    error_message = "crdb_locality value is not supported anymore. Use `locality` for similar behavior."
  }
}
