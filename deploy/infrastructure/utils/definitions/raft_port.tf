variable "raft_port" {
  type        = number
  description = <<-EOT
  Port of the Raft listener of the nodes of this instance.

  Only used with the raft datastore.

  Example: `2407`
  EOT

  default = 2407
}
