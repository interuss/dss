variable "evict_scd_subscriptions" {
  type        = bool
  description = <<-EOT
  Set this to true to enable cleanup of SCD subscriptions.

  EOT

  default = true
}
