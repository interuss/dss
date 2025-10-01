variable "evict_rid_isas" {
  type        = bool
  description = <<-EOT
  Set this to true to enable cleanup of RID ISAs.

  EOT

  default = true
}
