variable "kubernetes_namespace" {
  type        = string
  description = <<-EOT
  Namespace where to deploy Kubernetes resources.

  Example: `default`
  EOT

  default = "default"
}