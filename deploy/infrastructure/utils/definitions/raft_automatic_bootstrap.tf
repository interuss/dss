variable "raft_automatic_bootstrap" {
  type        = bool
  description = <<-EOT
  Form the Raft groups automatically: the first node of this instance bootstraps them, then the other nodes of
  this instance and the nodes of `raft_automatic_bootstrap_nodes` are added to them one at a time, retrying until
  all of them are members.

  Only meant for automation: enable it only for the first nodes of a cluster, otherwise independent Raft groups
  are created.

  Only used with the raft datastore.

  Example: `true`
  EOT

  default = false
}
