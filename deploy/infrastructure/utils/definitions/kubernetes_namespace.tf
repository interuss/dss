variable "kubernetes_namespace" {
  type        = string
  description = "Namespace where to deploy Kubernetes resources. Only default is supported at the moment."
  default     = "default"

  # TODO: Adapt current deployment scripts in /build/deploy to support default is supported for the moment.
  validation {
    condition     = var.kubernetes_namespace == "default"
    error_message = "Only default namespace is supported at the moment"
  }
}