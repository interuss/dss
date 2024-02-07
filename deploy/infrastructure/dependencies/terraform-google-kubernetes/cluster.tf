# Resources related to the kubernetes cluster

resource "google_container_cluster" "kubernetes_cluster" {
  name     = var.cluster_name
  location = var.google_zone

  remove_default_node_pool = true
  initial_node_count       = 1

  networking_mode = "VPC_NATIVE"
  ip_allocation_policy {
    # Intentionally left empty.
  }

  min_master_version = var.kubernetes_version
}

resource "google_container_node_pool" "dss_pool" {
  name       = "dss-pool"
  location   = var.google_zone
  cluster    = google_container_cluster.kubernetes_cluster.name
  node_count = var.node_count

  node_config {
    machine_type = var.google_machine_type
    disk_size_gb = 15 # Kubernetes PVC will create distinct disks.

    # TODO: Use non-default service account with IAM roles
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]

    metadata = {
      "disable-legacy-endpoints" = true
    }
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Static IP addresses for the gateway
resource "google_compute_global_address" "ip_gateway" {
  name       = format("%s-ip-gateway", var.cluster_name)
  ip_version = "IPV4"

  # Current google terraform provider doesn't allow tags or labels. Description is used to preserve mapping between ips and hostnames.
  description = var.app_hostname
}

# Static IP addresses for CRDB instances
resource "google_compute_address" "ip_crdb" {
  count  = var.node_count
  name   = format("%s-ip-crdb%v", var.cluster_name, count.index)
  region = local.region

  # Current google terraform provider doesn't allow tags or labels. Description is used to preserve mapping between ips and hostnames.
  description = format("%s.%s", count.index, var.crdb_hostname_suffix)
}

locals {
  kubectl_cluster_context_name = format("gke_%s_%s_%s", google_container_cluster.kubernetes_cluster.project, google_container_cluster.kubernetes_cluster.location, google_container_cluster.kubernetes_cluster.name)
}
