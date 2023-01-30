#output "crdb_addresses" {
#  value = module.terraform-aws-kubernetes.crdb_addresses
#}
#
#output "gateway_address" {
#  value = module.terraform-aws-kubernetes.gateway_address
#}

output "generated_files_location" {
  value = module.terraform-commons-dss.generated_files_location
}

output "cluster_context" {
  value = module.terraform-aws-kubernetes.kubernetes_context_name
}