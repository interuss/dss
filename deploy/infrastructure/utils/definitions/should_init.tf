variable "should_init" {
  type        = bool
  description = <<-EOT
  Set to false if joining an existing pool, true if creating the first DSS instance
  for a pool. When set true, this can initialize the data directories on your cluster,
  and prevent you from joining an existing pool.
  If should_init is true, crdb_external_nodes shall be empty

  Only used for CockroachDB

  Example: `true`
  EOT
}
