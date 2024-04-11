module "terraform-google-kubernetes" {
  # See variables.tf for variables description.
  google_project_name          = var.google_project_name
  cluster_name                 = var.cluster_name
  google_zone                  = var.google_zone
  app_hostname                 = var.app_hostname
  crdb_hostname_suffix         = var.crdb_hostname_suffix
  google_dns_managed_zone_name = var.google_dns_managed_zone_name
  google_machine_type          = var.google_machine_type
  node_count                   = var.node_count
  kubernetes_version           = var.kubernetes_version

  source = "../../dependencies/terraform-google-kubernetes"
}

module "terraform-commons-dss" {
  # See variables.tf for variables description.
  image                          = var.image
  kubernetes_namespace           = var.kubernetes_namespace
  kubernetes_storage_class       = var.google_kubernetes_storage_class
  app_hostname                   = var.app_hostname
  crdb_hostname_suffix           = var.crdb_hostname_suffix
  should_init                    = var.should_init
  authorization                  = var.authorization
  crdb_locality                  = var.crdb_locality
  image_pull_secret              = var.image_pull_secret
  crdb_external_nodes            = var.crdb_external_nodes
  kubernetes_api_endpoint        = module.terraform-google-kubernetes.kubernetes_api_endpoint
  crdb_internal_nodes            = module.terraform-google-kubernetes.crdb_nodes
  ip_gateway                     = module.terraform-google-kubernetes.ip_gateway
  ssl_policy                     = module.terraform-google-kubernetes.ssl_policy
  kubernetes_cloud_provider_name = module.terraform-google-kubernetes.kubernetes_cloud_provider_name
  kubernetes_context_name        = module.terraform-google-kubernetes.kubernetes_context_name
  kubernetes_get_credentials_cmd = module.terraform-google-kubernetes.kubernetes_get_credentials_cmd

  source = "../../dependencies/terraform-commons-dss"
}
