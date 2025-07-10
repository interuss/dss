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
  crdb_image_tag               = var.crdb_image_tag
  crdb_cluster_name            = var.crdb_cluster_name
  db_hostname_suffix           = var.db_hostname_suffix
  locality                     = var.locality
  crdb_external_nodes          = var.crdb_external_nodes
  image                        = var.image
  kubernetes_version           = var.kubernetes_version
  node_count                   = 3
  should_init                  = true
  enable_scd                   = true
}

