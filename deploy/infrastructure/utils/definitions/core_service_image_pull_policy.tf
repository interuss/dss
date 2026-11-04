variable "core_service_image_pull_policy" {
  type        = string
  default     = ""
  description = <<-EOT
  Kubernetes image pull policy of the core-service containers.
  When empty, `Always` is used (`IfNotPresent` on minikube with Tanka).

  Example: `IfNotPresent`
  EOT
}
