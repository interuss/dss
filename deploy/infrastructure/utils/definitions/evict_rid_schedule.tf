variable "evict_rid_schedule" {
  type        = string
  description = <<-EOT
  When the RID cleanup job shall be performed; expressed in cron format (https://crontab.guru/).

  EOT

  default = "*/30 * * * *"
}
