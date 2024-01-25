variable "kubernetes_version" {
  type        = string
  description = <<-EOT
    Desired version of the Kubernetes cluster control plane and nodes.

    Supported versions:
      - 1.24
  EOT

  validation {
    condition     = var.kubernetes_version == "1.24"
    error_message = "Only 1.24 is supported."
  }
}
