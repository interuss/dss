output "crdb_addresses" {
  value = [for a in google_compute_address.ip_crdb[*] : { expected_dns : a.description, address : a.address }]
}

output "gateway_address" {
  value = {
    expected_dns : google_compute_global_address.ip_gateway.description,
    address : google_compute_global_address.ip_gateway.address
  }
}

output "kubernetes_cloud_provider_name" {
  value = "google"
}

output "kubernetes_get_credentials_cmd" {
  value = "gcloud container clusters get-credentials --zone ${var.google_zone} ${var.cluster_name}"
}

output "kubernetes_api_endpoint" {
  value = google_container_cluster.kubernetes_cluster.endpoint
}

output "kubernetes_context_name" {
  value = local.kubectl_cluster_context_name
}

output "ip_gateway" {
  value = google_compute_global_address.ip_gateway.name
}

output "ssl_policy" {
  value = google_compute_ssl_policy.secure.name
}

output "crdb_nodes" {
  value = [
    for i in google_compute_address.ip_crdb : {
      ip  = i.address
      dns = i.description
    }
  ]
}
