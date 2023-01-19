variable "desired_scd_db_version" {
  type = string
  description = <<EOT
    Desired SCD DB schema version.
    Use `latest` to use the latest schema version.

    Example: `3.1.0`
  EOT

  default = "latest"
}