variable "evict_rid_ttl" {
  type        = string
  description = <<-EOT
  How long expired RID items should stay before being automatically removed; expressed in Go duration format (https://pkg.go.dev/time#ParseDuration).

  EOT

  default = "30m"
}
