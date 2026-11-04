variable "schema_manager_enabled" {
  type        = bool
  default     = true
  description = <<-EOT
  Set to false to not deploy the schema manager jobs.
  With Tanka, the schema manager jobs are always deployed when `should_init` is true.

  Example: `true`
  EOT
}
