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

output "crdb_nodes" {
  value = [
    for i in aws_eip.ip_crdb : {
      ip  = i.allocation_id
      dns = i.tags.ExpectedDNS
    }
  ]
  depends_on = [
    aws_eip.ip_crdb
  ]
}

output "yugabyte_masters_nodes" {
  value = [
    for i in aws_eip.ip_yugabyte_masters : {
      ip  = i.allocation_id
      dns = i.tags.ExpectedDNS
    }
  ]
  depends_on = [
    aws_eip.ip_yugabyte_masters
  ]
}

output "yugabyte_tservers_nodes" {
  value = [
    for i in aws_eip.ip_yugabyte_tservers : {
      ip  = i.allocation_id
      dns = i.tags.ExpectedDNS
    }
  ]
  depends_on = [
    aws_eip.ip_yugabyte_tservers
  ]
}

output "crdb_addresses" {
  value = [for i in aws_eip.ip_crdb[*] : { expected_dns : i.tags.ExpectedDNS, address : i.public_ip }]
}

output "yugabyte_masters_addresses" {
  value = [for i in aws_eip.ip_yugabyte_masters[*] : { expected_dns : i.tags.ExpectedDNS, address : i.public_ip }]
}

output "yugabyte_tservers_addresses" {
  value = [for i in aws_eip.ip_yugabyte_tservers[*] : { expected_dns : i.tags.ExpectedDNS, address : i.public_ip }]
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
