module "terraform-aws-kubernetes" {
  # See variables.tf for variables description.
  cluster_name                 = var.cluster_name
  aws_region                   = var.aws_region
  app_hostname                 = var.app_hostname
  crdb_hostname_suffix         = var.crdb_hostname_suffix
  aws_instance_type            = var.aws_instance_type
  aws_route53_zone_id          = var.aws_route53_zone_id
  aws_iam_permissions_boundary = var.aws_iam_permissions_boundary
  node_count                   = var.node_count

  source = "../../dependencies/terraform-aws-kubernetes"
}

module "terraform-commons-dss" {
  # See variables.tf for variables description.
  image                          = var.image
  image_pull_secret              = var.image_pull_secret
  kubernetes_namespace           = var.kubernetes_namespace
  kubernetes_storage_class       = var.aws_kubernetes_storage_class
  app_hostname                   = var.app_hostname
  crdb_hostname_suffix           = var.crdb_hostname_suffix
  should_init                    = var.should_init
  authorization                  = var.authorization
  crdb_locality                  = var.crdb_locality
  crdb_internal_nodes            = module.terraform-aws-kubernetes.crdb_nodes
  ip_gateway                     = module.terraform-aws-kubernetes.ip_gateway
  kubernetes_api_endpoint        = module.terraform-aws-kubernetes.kubernetes_api_endpoint
  kubernetes_cloud_provider_name = module.terraform-aws-kubernetes.kubernetes_cloud_provider_name
  kubernetes_context_name        = module.terraform-aws-kubernetes.kubernetes_context_name
  kubernetes_get_credentials_cmd = module.terraform-aws-kubernetes.kubernetes_get_credentials_cmd
  workload_subnet                = module.terraform-aws-kubernetes.workload_subnet
  gateway_cert_name              = module.terraform-aws-kubernetes.app_hostname_cert_arn

  source = "../../dependencies/terraform-commons-dss"
}
