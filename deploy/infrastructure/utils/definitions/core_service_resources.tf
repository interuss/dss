variable "core_service_resources" {
  type        = map(map(string))
  description = <<-EOT
  Kubernetes CPU and memory requests and limits of the core-service containers.
  Leave empty to not set any.

  Example: `{ requests = { cpu = "0.5", memory = "2Gi" }, limits = { memory = "2Gi" } }`
  EOT

  default = {}

  validation {
    condition     = alltrue([for k in keys(var.core_service_resources) : contains(["requests", "limits"], k)])
    error_message = "core_service_resources keys must be `requests` or `limits`."
  }
}
