variable "evict_scd_batch_size" {
  type        = number
  description = <<-EOT
  Number of entities deleted per batch query (per type) by each run of the SCD eviction command.
  Set to 0 to delete all expired entities in a single query (per type).

  Example: `100`
  EOT

  default = 0

  validation {
    condition     = var.evict_scd_batch_size >= 0
    error_message = "evict_scd_batch_size must be equal to or greater than 0."
  }
}
