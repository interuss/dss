variable "aws_route53_zone_id" {
  type        = string
  description = <<-EOT
    AWS Route 53 Zone ID
    This module can automatically create DNS records in a Route 53 Zone.
    Leave empty to disable record creation.

    Example: `Z0123456789ABCDEFGHIJ`
  EOT
}