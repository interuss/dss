variable "crdb_image_tag" {
  type        = string
  description = <<-EOT
    Version tag of the CockroachDB image.
    Until v.16, the recommended CockroachDB version is v21.2.7.
    From v.17, the recommended CockroachDB version is v24.1.3.

    Example: v24.1.3
  EOT
}
