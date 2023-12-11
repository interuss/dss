
resource "aws_eks_addon" "aws-ebs-csi-driver" {
  addon_name               = "aws-ebs-csi-driver"
  cluster_name             = aws_eks_cluster.kubernetes_cluster.name
  service_account_role_arn = aws_iam_role.AmazonEKS_EBS_CSI_DriverRole.arn
  depends_on = [
    aws_eks_node_group.eks_node_group
  ]
}
