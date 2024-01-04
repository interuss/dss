terraform {
  backend "s3" {
    bucket = "interuss-tf-backend-ci"
    key    = "aws-1"
    region = "us-east-1"
  }
}

module "terraform-aws-dss" {
  source = "../../../infrastructure/modules/terraform-aws-dss"

  app_hostname                 = var.app_hostname
  authorization                = var.authorization
  aws_iam_permissions_boundary = var.aws_iam_permissions_boundary
  aws_instance_type            = var.aws_instance_type
  aws_kubernetes_storage_class = var.aws_kubernetes_storage_class
  aws_region                   = var.aws_region
  aws_route53_zone_id          = var.aws_route53_zone_id
  cluster_name                 = var.cluster_name
  crdb_hostname_suffix         = var.crdb_hostname_suffix
  crdb_locality                = var.crdb_locality
  image                        = var.image
  node_count                   = 3
  should_init                  = true
  enable_scd                   = true
}

