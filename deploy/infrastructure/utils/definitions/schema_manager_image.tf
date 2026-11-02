variable "schema_manager_image" {
  type        = string
  default     = ""
  description = <<-EOT
  Image used by the schema manager and evict jobs. When empty, `image` is used.

  Example: `docker.io/interuss/dss:v0.15.0`
  EOT
}
