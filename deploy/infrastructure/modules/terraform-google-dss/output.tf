output "crdb_addresses" {
  value = module.terraform-google-kubernetes.crdb_addresses
}

output "gateway_address" {
  value = module.terraform-google-kubernetes.gateway_address
}

output "generated_files_location" {
  value = module.terraform-commons-dss.generated_files_location
}