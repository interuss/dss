variable "aws_region" {
  type        = string
  description = <<-EOT
    AWS region
    List of available regions: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-regions
    Currently, the terraform module uses the two first availability zones of the region.

    Example: `eu-west-1`
  EOT
}