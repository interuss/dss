# Resources related to the kubernetes cluster

resource "google_container_cluster" "kubernetes_cluster" {
  name     = var.google_cluster_context.name
  location = var.google_cluster_context.zone

  remove_default_node_pool = true
  initial_node_count       = 1

  networking_mode = "VPC_NATIVE"
  ip_allocation_policy {
    # Intentionally left empty.
  }
}

resource "google_container_node_pool" "dss_pool" {
  name       = "dss-pool"
  location   = var.google_cluster_context.zone
  cluster    = google_container_cluster.kubernetes_cluster.name
  node_count = var.dss_configuration.crdb_node_count

  node_config {
    machine_type = var.google_cluster_context.machine_type

    # TODO: Use non-default service account with IAM roles
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Static IP addresses for the gateway
resource "google_compute_global_address" "ip_gateway" {
  name       = format("%s-ip-gateway", var.google_cluster_context.name)
  ip_version = "IPV4"

  # Current google terraform provider doesn't allow tags or labels. Description is used to preserve mapping between ips and hostnames.
  description = var.dss_configuration.app_hostname
}

# Static IP addresses for CRDB instances
resource "google_compute_address" "ip_crdb" {
  count  = var.dss_configuration.crdb_node_count
  name   = format("%s-ip-crdb%v", var.google_cluster_context.name, count.index)
  region = var.google_cluster_context.region

  # Current google terraform provider doesn't allow tags or labels. Description is used to preserve mapping between ips and hostnames.
  description = format("%s.%s", count.index, var.dss_configuration.crdb_hostname_suffix)
}

locals {
  kubectl_cluster_context_name = format("gke_%s_%s_%s", google_container_cluster.kubernetes_cluster.project, google_container_cluster.kubernetes_cluster.location, google_container_cluster.kubernetes_cluster.name)
}