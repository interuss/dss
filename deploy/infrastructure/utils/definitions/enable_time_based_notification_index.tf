variable "enable_time_based_notification_index" {
  type        = bool
  description = "Set this boolean to true to use a time-based notification index when working with RID and SCD subscriptions. Must be enabled on all instances part of the pool."
  default     = false
}
