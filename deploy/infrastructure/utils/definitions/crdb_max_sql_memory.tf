variable "crdb_max_sql_memory" {
  type        = string
  description = <<-EOT
  Maximum memory CockroachDB can use for SQL queries, passed to the `--max-sql-memory` flag of `cockroach start`.
  Either a percentage of the memory available to the container or an absolute size.

  Example: `25%` or `2GiB`
  EOT

  default = "25%"
}
