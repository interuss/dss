resource "aws_eks_addon" "node-monitoring-agent" {
  count = var.aws_enable_node_monitoring_agent ? 1 : 0

  addon_name   = "eks-node-monitoring-agent"
  cluster_name = aws_eks_cluster.kubernetes_cluster.name
  depends_on = [
    aws_eks_node_group.eks_node_group
  ]
}
