output "kubernetes_cloud_provider_name" {
  value = "aws"
}

output "kubernetes_get_credentials_cmd" {
  value = "aws eks --region ${var.aws_region} update-kubeconfig --name ${var.cluster_name}"
}

output "kubernetes_api_endpoint" {
  value = aws_eks_cluster.kubernetes_cluster.endpoint
}

output "kubernetes_context_name" {
  value = aws_eks_cluster.kubernetes_cluster.arn
}

output "ip_gateway" {
  value = aws_eip.gateway[0].id
}

output "ip_prometheus" {
  value = length(aws_eip.ip_prometheus) > 0 ? aws_eip.ip_prometheus[0].id : ""
}

output "crdb_addresses" {
  value = [for i in aws_eip.ip_crdb[*] : { expected_dns : i.tags.ExpectedDNS, address : i.public_ip, loadbalancer_ip : i.allocation_id }]
}

output "yugabyte_masters_addresses" {
  value = [for i in aws_eip.ip_yugabyte[*] : { expected_dns : i.tags.ExpectedMasterDNS, address : i.public_ip, loadbalancer_ip : i.allocation_id }]
}

output "yugabyte_tservers_addresses" {
  value = [for i in aws_eip.ip_yugabyte[*] : { expected_dns : i.tags.ExpectedTServerDNS, address : i.public_ip, loadbalancer_ip : i.allocation_id }]
}

output "raft_addresses" {
  value = [for i in aws_eip.ip_raft[*] : { expected_dns : i.tags.ExpectedDNS, address : i.public_ip, loadbalancer_ip : i.allocation_id }]
}

output "gateway_address" {
  value = {
    expected_dns : aws_eip.gateway[0].tags.ExpectedDNS,
    address : aws_eip.gateway[0].public_ip,
    certificate_validation_dns : [
      for c in aws_acm_certificate.app_hostname.domain_validation_options[*] : {
        managed_by_terraform : length(aws_route53_record.app_hostname_cert_validation) > 0
        name : c.resource_record_name,
        type : c.resource_record_type,
        records : [
          c.resource_record_value
        ]
    }]
  }
}

output "workload_subnet" {
  value = data.aws_subnet.main_subnet.id
}

output "iam_role_node_group_arn" {
  value = aws_iam_role.dss-cluster-node-group.arn
}
