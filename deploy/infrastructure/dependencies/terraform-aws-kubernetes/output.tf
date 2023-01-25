output "kubernetes_cloud_provider_name" {
  value = "google"
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

#output "ip_gateway" {
#  value = ...
#}
#
#output "crdb_nodes" {
#  value = [
#    for i in ... : {
#      ip  = i.address
#      dns = i.description
#    }
#  ]
#}