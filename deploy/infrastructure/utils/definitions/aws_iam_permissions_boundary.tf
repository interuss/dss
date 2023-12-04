variable "aws_iam_permissions_boundary" {
  type        = string
  description = <<-EOT
    AWS IAM Policy to be used for permissions boundaries on created roles.

    Example: `GithubCIPermissionBoundaries`
  EOT
}