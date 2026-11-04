variable "evict_scd_schedule" {
  type        = string
  description = <<-EOT
  When the SCD cleanup job shall be performed; expressed in cron format (https://crontab.guru/).
  Leave empty to run daily at 2 AM at a minute deterministically derived from the locality.

  EOT

  default = ""
}
