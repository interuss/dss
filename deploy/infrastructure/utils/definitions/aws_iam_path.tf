variable "aws_iam_path" {
  type        = string
  description = <<-EOT
    AWS IAM Resources Path
    IAM related resources will be created within the specified path

    Example: `ci/`
  EOT
  default     = ""
}