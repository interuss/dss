variable "raft_automatic_bootstrap_nodes" {
  type        = list(string)
  description = <<-EOT
  Addresses (host:port, no scheme) of the Raft listeners of the nodes of other participants to add to the Raft
  groups with `raft_automatic_bootstrap`.

  Only used with the raft datastore.

  Example: `["0.raft.dss.example.com:2407", "1.raft.dss.example.com:2407"]`
  EOT

  default = []
}
