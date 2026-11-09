variable "raft_tls_secret" {
  type        = string
  description = <<-EOT
  Name of the Kubernetes secret holding the mutual TLS material of the Raft listeners:
  `node.raft-<i>.crt`/`node.raft-<i>.key` is the certificate of the pod of ordinal `<i>` of this instance,
  `ca.crt` is the PEM bundle of the CAs of every participant of the pool, including its own.

  Only used with the raft datastore.

  Example: `dss.raft.certs`
  EOT

  default = "dss.raft.certs"
}
