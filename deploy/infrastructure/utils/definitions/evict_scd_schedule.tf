variable "evict_scd_schedule" {
  type        = string
  description = <<-EOT
  When the SCD cleanup job shall be performed; expressed in cron format (https://crontab.guru/).

  EOT

  default = "0 2 * * *"
}
