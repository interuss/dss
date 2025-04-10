variable "aws_instance_type" {
  type        = string
  description = <<-EOT
  AWS EC2 instance type used for the Kubernetes node pool.

  Example: `m6g.xlarge` for production and `t3.medium` for development
  EOT
}