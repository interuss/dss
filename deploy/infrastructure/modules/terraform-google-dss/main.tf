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

  source = "../../dependencies/terraform-google-kubernetes"
}

module "terraform-commons-dss" {
  # See variables.tf for variables description.
  kubernetes_namespace           = var.kubernetes_namespace
  kubernetes_storage_class       = var.kubernetes_storage_class
  app_hostname                   = var.app_hostname
  crdb_hostname_suffix           = var.crdb_hostname_suffix
  should_init                    = var.should_init
  authorization                  = var.authorization
  crdb_locality                  = var.crdb_locality
  kubernetes_api_endpoint        = module.terraform-google-kubernetes.kubernetes_api_endpoint
  crdb_internal_nodes            = module.terraform-google-kubernetes.crdb_nodes
  crdb_internal_addresses        = module.terraform-google-kubernetes.internal_node_addresses
  ip_gateway                     = module.terraform-google-kubernetes.ip_gateway
  kubernetes_cloud_provider_name = module.terraform-google-kubernetes.kubernetes_cloud_provider_name
  kubernetes_context_name        = module.terraform-google-kubernetes.kubernetes_context_name
  kubernetes_get_credentials_cmd = module.terraform-google-kubernetes.kubernetes_get_credentials_cmd

  source = "../../dependencies/terraform-commons-dss"
}
