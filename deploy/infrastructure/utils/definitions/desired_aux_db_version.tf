variable "desired_aux_db_version" {
  type        = string
  description = <<-EOT
  Desired AUX DB schema version.
  Use `latest` to use the latest schema version.

  Example: `3.1.0`
  EOT

  default = "latest"
}
