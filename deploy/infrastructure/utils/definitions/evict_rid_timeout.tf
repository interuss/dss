variable "evict_rid_timeout" {
  type        = string
  description = <<-EOT
  Timeout of the RID eviction command; expressed in Go duration format (https://pkg.go.dev/time#ParseDuration).
  Leave empty to use the default value of the command.

  Example: `10m`
  EOT

  default = ""
}
