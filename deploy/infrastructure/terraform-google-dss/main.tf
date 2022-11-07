# Providers
provider "google" {
  region  = var.google_cluster_context.region
  project = var.google_cluster_context.project
}

module "terraform-commons-dss" {
  source = "../terraform-commons-dss"

  dss_configuration = var.dss_configuration

  kubernetes = {
    provider_name                = "google"
    get_credentials_cmd          = "gcloud container clusters get-credentials --zone ${var.google_cluster_context.zone} ${var.google_cluster_context.name}"
    api_endpoint                 = google_container_cluster.kubernetes_cluster.endpoint
    kubectl_cluster_context_name = local.kubectl_cluster_context_name
    node_addresses               = google_compute_address.ip_crdb[*].description
    ip_gateway                   = google_compute_global_address.ip_gateway.name
    crdb_nodes = [for i in google_compute_address.ip_crdb : {
      ip  = i.address
      dns = i.description
    }]
  }
}