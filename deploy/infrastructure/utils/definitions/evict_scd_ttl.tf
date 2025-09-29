variable "evict_scd_ttl" {
  type        = string
  description = <<-EOT
  How long expired SCD items should stay before being automatically removed; expressed in Go duration format (https://pkg.go.dev/time#ParseDuration).

  EOT

  default = "2688h"
}
