variable "crdb_env" {
  type        = map(string)
  description = <<-EOT
  Additional environment variables of the CockroachDB containers, for instance to tune the Go runtime.

  Example: `{ GOGC = "80", GOMAXPROCS = "2" }`
  EOT

  default = {}
}
