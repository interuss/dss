variable "raft_peers" {
  type = list(object({
    id       = number
    endpoint = string
  }))
  description = <<-EOT
  The other nodes of the Raft pool, deployed by other instances.
  `endpoint` is the address (host:port, without scheme) of the node, reachable from this cluster.

  Only used with the raft datastore.

  Example: `[{ id = 4, endpoint = "0.db.dss.example.com:2407" }]`
  EOT
  default     = []
}
