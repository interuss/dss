variable "evict_rid_subscriptions" {
  type        = bool
  description = <<-EOT
  Set this to true to enable cleanup of RID subscriptions.

  EOT

  default = true
}
