variable "db_hostname_suffix" {
  type        = string
  description = <<-EOT
  The domain name suffix shared by all of your databases nodes.
  For instance, if your database nodes were addressable at 0.db.example.com,
  1.db.example.com and 2.db.example.com (CockroachDB) or 0.master.db.example.com, 1.tserver.db.example.com (Yugabyte), then the value would be db.example.com.

  Example: db.example.com
  EOT
}
