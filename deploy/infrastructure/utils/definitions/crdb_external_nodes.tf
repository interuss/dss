variable "crdb_external_nodes" {
  type        = list(string)
  description = <<-EOT
    Fully-qualified domain name of existing CRDB nodes outside of the cluster if you are joining an existing pool.
  EOT
  default     = []
}