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