output "crdb_addresses" {
  value = module.terraform-google-kubernetes.crdb_addresses
}

output "yugabyte_masters_addresses" {
  value = module.terraform-google-kubernetes.yugabyte_masters_addresses
}

output "yugabyte_tservers_addresses" {
  value = module.terraform-google-kubernetes.yugabyte_tservers_addresses
}

output "gateway_address" {
  value = module.terraform-google-kubernetes.gateway_address
}

output "generated_files_location" {
  value = module.terraform-commons-dss.generated_files_location
}

output "workspace_location" {
  value = module.terraform-commons-dss.workspace_location
}

output "cluster_context" {
  value = module.terraform-google-kubernetes.kubernetes_context_name
}

