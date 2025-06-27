variable "should_init" {
  type        = bool
  description = <<-EOT
  Set to false if joining an existing pool, true if creating the first DSS instance
  for a pool. When set true, this can initialize the data directories on your cluster,
  and prevent you from joining an existing pool.

  Only used for CockroachDB with Tanka

  Example: `true`
  EOT
}
