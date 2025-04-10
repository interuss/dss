variable "aws_iam_permissions_boundary" {
  type        = string
  description = <<-EOT
  AWS IAM Policy ARN to be used for permissions boundaries on created roles.

  Example: `arn:aws:iam::123456789012:policy/GithubCIPermissionBoundaries`
  EOT

  default = ""
}
