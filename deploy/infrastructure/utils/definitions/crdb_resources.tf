variable "crdb_resources" {
  type        = map(map(string))
  description = <<-EOT
  Kubernetes CPU and memory requests and limits of the CockroachDB containers.
  Leave empty to not set any.

  Example: `{ requests = { cpu = "2", memory = "10Gi" }, limits = { cpu = "2", memory = "10Gi" } }`
  EOT

  default = {}

  validation {
    condition     = alltrue([for k in keys(var.crdb_resources) : contains(["requests", "limits"], k)])
    error_message = "crdb_resources keys must be `requests` or `limits`."
  }
}
