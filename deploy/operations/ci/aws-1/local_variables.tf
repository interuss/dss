# This file contains variables only used by this module and which are not provided to child dependencies.

variable "aws_iam_administrator_role" {
  type        = string
  description = <<-EOT
    AWS IAM administrator role
    ARN of the role assumed by administrators when login into the AWS InterUSS account.

    Example: `arn:aws:iam::123456789012:role/AdminRole`
  EOT
}

variable "aws_iam_ci_role" {
  type        = string
  description = <<-EOT
    AWS IAM administrator role
    ARN of the role assumed by administrators when login into the AWS InterUSS account.

    Example: `arn:aws:iam::123456789012:role/CiRole`
  EOT
}
