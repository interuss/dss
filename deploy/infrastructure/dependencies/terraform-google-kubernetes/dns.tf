# Resources related to the DNS entries if managed. (See google_cluster.dns_managed_zone_name configuration field for details)
data "google_dns_managed_zone" "default" {
  name  = var.google_dns_managed_zone_name
  count = var.google_dns_managed_zone_name == "" ? 0 : 1
}

resource "google_dns_record_set" "gateway" {
  count = var.google_dns_managed_zone_name == "" ? 0 : 1
  name  = "${google_compute_global_address.ip_gateway.description}." # description contains the expected hostname
  type  = "A"
  ttl   = 300

  managed_zone = data.google_dns_managed_zone.default[0].name
  rrdatas      = [google_compute_global_address.ip_gateway.address]

}

resource "google_dns_record_set" "crdb" {
  count = var.datastore_type == "cockroachdb" && var.google_dns_managed_zone_name != "" ? var.node_count : 0
  name  = "${google_compute_address.ip_crdb[count.index].description}." # description contains the expected hostname
  type  = "A"
  ttl   = 300

  managed_zone = data.google_dns_managed_zone.default[0].name
  rrdatas      = [google_compute_address.ip_crdb[count.index].address]
}

resource "google_dns_record_set" "yugabyte_masters" {
  count = var.datastore_type == "yugabyte" && var.google_dns_managed_zone_name != "" ? var.node_count : 0
  name  = format("${google_compute_address.ip_yugabyte[count.index].description}.", "master") # description contains the expected hostname
  type  = "A"
  ttl   = 300

  managed_zone = data.google_dns_managed_zone.default[0].name
  rrdatas      = [google_compute_address.ip_yugabyte[count.index].address]
}

resource "google_dns_record_set" "yugabyte_tserver" {
  count = var.datastore_type == "yugabyte" && var.google_dns_managed_zone_name != "" ? var.node_count : 0
  name  = format("${google_compute_address.ip_yugabyte[count.index].description}.", "tserver") # description contains the expected hostname
  type  = "A"
  ttl   = 300

  managed_zone = data.google_dns_managed_zone.default[0].name
  rrdatas      = [google_compute_address.ip_yugabyte[count.index].address]
}
