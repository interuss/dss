output "crdb_addresses" {
  value = [for a in google_compute_address.ip_crdb[*] : { expected_dns : a.description, address : a.address }]
}

output "gateway_address" {
  value = { expected_dns : google_compute_global_address.ip_gateway.description, address : google_compute_global_address.ip_gateway.address }
}