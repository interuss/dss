variable "psp_role_ref" {
  type        = string
  default     = ""
  description = <<-EOT
  Name of the ClusterRole bound to all the service accounts of the namespace, for Pod Security Policies.
  No RoleBinding is created when empty.

  Example: `psp:privileged`
  EOT
}
