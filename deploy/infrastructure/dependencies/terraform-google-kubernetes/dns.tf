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
  count = var.google_dns_managed_zone_name == "" || var.datastore_type != "cockroachdb" ? 0 : var.node_count
  name  = "${google_compute_address.ip_crdb[count.index].description}." # description contains the expected hostname
  type  = "A"
  ttl   = 300

  managed_zone = data.google_dns_managed_zone.default[0].name
  rrdatas      = [google_compute_address.ip_crdb[count.index].address]
}

resource "google_dns_record_set" "yugabyte_masters" {
  count = var.google_dns_managed_zone_name == "" || var.datastore_type != "yugabyte" ? 0 : var.node_count
  name  = "${google_compute_address.ip_yugabyte_masters[count.index].description}." # description contains the expected hostname
  type  = "A"
  ttl   = 300

  managed_zone = data.google_dns_managed_zone.default[0].name
  rrdatas      = [google_compute_address.ip_yugabyte_masters[count.index].address]
}

resource "google_dns_record_set" "yugabyte_tserver" {
  count = var.google_dns_managed_zone_name == "" || var.datastore_type != "yugabyte" ? 0 : var.node_count
  name  = "${google_compute_address.ip_yugabyte_tservers[count.index].description}." # description contains the expected hostname
  type  = "A"
  ttl   = 300

  managed_zone = data.google_dns_managed_zone.default[0].name
  rrdatas      = [google_compute_address.ip_yugabyte_tservers[count.index].address]
}
