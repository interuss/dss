module "terraform-aws-kubernetes" {
  # See variables.tf for variables description.
  cluster_name                 = var.cluster_name
  aws_region                   = var.aws_region
  app_hostname                 = var.app_hostname
  db_hostname_suffix           = var.db_hostname_suffix
  datastore_type               = var.datastore_type
  aws_instance_type            = var.aws_instance_type
  aws_route53_zone_id          = var.aws_route53_zone_id
  aws_iam_permissions_boundary = var.aws_iam_permissions_boundary
  node_count                   = var.node_count
  kubernetes_version           = var.kubernetes_version

  source = "../../dependencies/terraform-aws-kubernetes"
}

module "terraform-commons-dss" {
  # See variables.tf for variables description.
  image                            = var.image
  image_pull_secret                = var.image_pull_secret
  kubernetes_namespace             = var.kubernetes_namespace
  kubernetes_storage_class         = var.aws_kubernetes_storage_class
  app_hostname                     = var.app_hostname
  crdb_image_tag                   = var.crdb_image_tag
  crdb_cluster_name                = var.crdb_cluster_name
  db_hostname_suffix               = var.db_hostname_suffix
  datastore_type                   = var.datastore_type
  should_init                      = var.should_init
  authorization                    = var.authorization
  locality                         = var.locality
  crdb_external_nodes              = var.crdb_external_nodes
  node_count                       = var.node_count
  yugabyte_cloud                   = var.yugabyte_cloud
  yugabyte_region                  = var.yugabyte_region
  yugabyte_zone                    = var.yugabyte_zone
  yugabyte_light_resources         = var.yugabyte_light_resources
  yugabyte_external_nodes          = var.yugabyte_external_nodes
  crdb_internal_nodes              = module.terraform-aws-kubernetes.crdb_nodes
  yugabyte_internal_masters_nodes  = module.terraform-aws-kubernetes.yugabyte_masters_nodes
  yugabyte_internal_tservers_nodes = module.terraform-aws-kubernetes.yugabyte_tservers_nodes
  ip_gateway                       = module.terraform-aws-kubernetes.ip_gateway
  kubernetes_api_endpoint          = module.terraform-aws-kubernetes.kubernetes_api_endpoint
  kubernetes_cloud_provider_name   = module.terraform-aws-kubernetes.kubernetes_cloud_provider_name
  kubernetes_context_name          = module.terraform-aws-kubernetes.kubernetes_context_name
  kubernetes_get_credentials_cmd   = module.terraform-aws-kubernetes.kubernetes_get_credentials_cmd
  workload_subnet                  = module.terraform-aws-kubernetes.workload_subnet
  gateway_cert_name                = module.terraform-aws-kubernetes.app_hostname_cert_arn

  source = "../../dependencies/terraform-commons-dss"
}
