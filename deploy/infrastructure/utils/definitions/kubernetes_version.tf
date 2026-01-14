variable "kubernetes_version" {
  type        = string
  description = <<-EOT
  Desired version of the Kubernetes cluster control plane and nodes.

  Supported versions: 1.28 to 1.34
  EOT

  validation {
    condition     = contains(["1.28", "1.29", "1.30", "1.31", "1.32", "1.33", "1.34"], var.kubernetes_version)
    error_message = "Supported versions: 1.28 to 1.34"
  }
}
