variable "aws_instance_type" {
  type        = string
  description = <<-EOT
  AWS EC2 instance type used for the Kubernetes node pool.
  See https://aws.amazon.com/ec2/instance-types/ for available options.

  Depending on your use case, performance may be significantly improved with higher-tier instances, though this should be balanced against the associated costs.

  Both CockroachDB and YugabyteDB recommend `m6i` instances for production. Use `t` instances for testing only.

  See https://www.cockroachlabs.com/docs/v24.1/recommended-production-settings#aws and https://docs.yugabyte.com/stable/deploy/checklist/#amazon-web-services-aws for database-specific recommendations.

  Example: `m6i.xlarge` for production and `t3.medium` for development.
  EOT
}
