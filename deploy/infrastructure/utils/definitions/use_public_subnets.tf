variable "use_public_subnets" {
  type        = bool
  default     = true
  description = <<-EOT
  Set to true to use public subnets for the Kubernetes cluster. It is recommended to use private subnets for production environments for security reasons.
  For backward compatibility, the default is true.

  Example: `true`
  EOT
}
