resource "aws_eks_addon" "metrics-server" {
  addon_name   = "metrics-server"
  cluster_name = aws_eks_cluster.kubernetes_cluster.name
  depends_on = [
    aws_eks_node_group.eks_node_group
  ]
}
