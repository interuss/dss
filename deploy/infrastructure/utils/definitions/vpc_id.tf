variable "vpc_id" {
  type        = string
  default     = ""
  description = <<-EOT
  ID of the VPC to use for the Kubernetes cluster, instead of creating a new one.
  For backward compatibility, the default is blank.

  Example: `vpc-01234567890123456`
  EOT
}
