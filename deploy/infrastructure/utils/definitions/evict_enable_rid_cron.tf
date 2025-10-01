variable "evict_enable_rid_cron" {
  type        = bool
  description = <<-EOT
  Set this to true to enable the cron job that automatically cleanup RID entries.

  EOT

  default = true
}
