variable "desired_rid_db_version" {
  type = string
  description = <<EOT
    Desired RID DB schema version.
    Use `latest` to use the latest schema version.

    Example: `4.0.0`
  EOT

  default = "latest"
}