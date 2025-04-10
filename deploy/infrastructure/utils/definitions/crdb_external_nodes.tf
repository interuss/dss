variable "crdb_external_nodes" {
  type        = list(string)
  description = <<-EOT
  Fully-qualified domain name of existing CRDB nodes outside of the cluster if you are joining an existing pool.
  Example: ["0.db.dss.example.com", "1.db.dss.example.com", "2.db.dss.example.com"]
  EOT
  default     = []
}