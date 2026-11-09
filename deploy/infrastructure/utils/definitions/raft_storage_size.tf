variable "raft_storage_size" {
  type        = string
  description = <<-EOT
  Size of the persistent volume of each Raft node.

  Only used with the raft datastore.

  Example: `100Gi`
  EOT

  default = "100Gi"
}
