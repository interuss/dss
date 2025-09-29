variable "evict_enable_scd_cron" {
  type        = bool
  description = <<-EOT
  Set this to true to enable the cron job that automatically cleanup expired SCD entries.

  EOT

  default = false
}
