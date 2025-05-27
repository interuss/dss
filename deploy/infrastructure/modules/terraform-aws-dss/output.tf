output "crdb_addresses" {
  value = module.terraform-aws-kubernetes.crdb_addresses
}

output "yugabyte_masters_addresses" {
  value = module.terraform-aws-kubernetes.yugabyte_masters_addresses
}

output "yugabyte_tservers_addresses" {
  value = module.terraform-aws-kubernetes.yugabyte_tservers_addresses
}

output "gateway_address" {
  value = module.terraform-aws-kubernetes.gateway_address
}

output "iam_role_node_group_arn" {
  value = module.terraform-aws-kubernetes.iam_role_node_group_arn
}

output "generated_files_location" {
  value = module.terraform-commons-dss.generated_files_location
}

output "workspace_location" {
  value = module.terraform-commons-dss.workspace_location
}

output "cluster_context" {
  value = module.terraform-aws-kubernetes.kubernetes_context_name
}

