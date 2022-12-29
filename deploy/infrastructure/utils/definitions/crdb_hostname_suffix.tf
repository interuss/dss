variable "crdb_hostname_suffix" {
  type = string
  description = <<-EOT
  The domain name suffix shared by all of your CockroachDB nodes.
  For instance, if your CRDB nodes were addressable at 0.db.example.com,
  1.db.example.com and 2.db.example.com, then VAR_CRDB_HOSTNAME_SUFFIX would be db.example.com.
  Example: db.example.com
  EOT
}