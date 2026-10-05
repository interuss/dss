variable "evict_rid_limit" {
  type        = number
  description = <<-EOT
  Maximum number of entities deleted by each run of the RID eviction command.
  Set to 0 to delete all expired entities in a single run.

  Example: `1000`
  EOT

  default = 0

  validation {
    condition     = var.evict_rid_limit >= 0
    error_message = "evict_rid_limit must be equal to or greater than 0."
  }
}
