variable "yugabyte_external_nodes" {
  type        = list(string)
  description = <<-EOT
  Fully-qualified domain name of existing yugabyte master nodes outside of the cluster if you are joining an existing pool.
  Example: ["0.master.db.dss.example.com", "1.master.db.dss.example.com", "2.master.db.dss.example.com"]
  EOT
  default     = []
}

