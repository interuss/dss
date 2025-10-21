variable "evict_scd_operational_intents" {
  type        = bool
  description = <<-EOT
  Set this to true to enable cleanup of SCD operational intents.

  EOT

  default = true
}
