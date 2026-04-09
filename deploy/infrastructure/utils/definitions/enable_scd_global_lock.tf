variable "enable_scd_global_lock" {
  type        = bool
  description = "Set this boolean true to enable experimental global lock when working with SCD subscriptions. Reduce global throughput but improve throughput with lot of subscriptions in the same areas."
  default     = false
}
