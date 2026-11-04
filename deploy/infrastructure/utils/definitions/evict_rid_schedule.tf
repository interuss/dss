variable "evict_rid_schedule" {
  type        = string
  description = <<-EOT
  When the RID cleanup job shall be performed; expressed in cron format (https://crontab.guru/).
  Leave empty to run hourly at a minute deterministically derived from the locality.

  EOT

  default = ""
}
