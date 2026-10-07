variable "crdb_cache" {
  type        = string
  description = <<-EOT
  Size of the CockroachDB storage engine cache, passed to the `--cache` flag of `cockroach start`.
  Either a percentage of the memory available to the container or an absolute size.

  Example: `25%` or `2GiB`
  EOT

  default = "25%"
}
