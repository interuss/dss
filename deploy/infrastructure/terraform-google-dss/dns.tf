# Resources related to the DNS entries if managed. (See google_cluster_context.dns_managed_zone_name configuration field for details)
data "google_dns_managed_zone" "default" {
  name = var.google_cluster_context.dns_managed_zone_name
}

resource "google_dns_record_set" "gateway" {
  count = var.google_cluster_context.dns_managed_zone_name == "" ? 0 : 1
  name  = "${google_compute_global_address.ip_gateway.description}." # description contains the expected hostname
  type  = "A"
  ttl   = 300

  managed_zone = data.google_dns_managed_zone.default.name
  rrdatas      = [google_compute_global_address.ip_gateway.address]

}

resource "google_dns_record_set" "crdb" {
  count = var.google_cluster_context.dns_managed_zone_name == "" ? 0 : var.dss_configuration.crdb_node_count
  name  = "${google_compute_address.ip_crdb[count.index].description}." # description contains the expected hostname
  type  = "A"
  ttl   = 300

  managed_zone = data.google_dns_managed_zone.default.name
  rrdatas      = [google_compute_address.ip_crdb[count.index].address]
}