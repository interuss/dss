resource "aws_eks_cluster" "kubernetes_cluster" {
  name     = var.cluster_name
  role_arn = aws_iam_role.dss-cluster.arn

  vpc_config {
    subnet_ids             = local.vpc_id == "" ? aws_subnet.dss[*].id : var.private_subnet_ids
    endpoint_private_access = var.use_public_subnets ? false : true
    endpoint_public_access = var.use_public_subnets ? true : false
    public_access_cidrs = local.public_access_cidrs
    security_group_ids = [aws_security_group.eks-controlplane.id]
  }

  # Ensure that IAM Role permissions are created before and deleted after EKS Cluster handling.
  # Otherwise, EKS will not be able to properly delete EKS managed EC2 infrastructure such as Security Groups.
  depends_on = [
    aws_iam_role.dss-cluster-node-group,
    aws_iam_role_policy_attachment.dss-cluster-service,
    aws_iam_role_policy_attachment.AmazonEKSWorkerNodePolicy,
    aws_iam_role_policy_attachment.AmazonEKS_CNI_Policy,
    aws_internet_gateway.dss,
    aws_eip.gateway,
    aws_eip.ip_crdb,
    aws_security_group.eks-controlplane
  ]

  version = var.kubernetes_version
}

resource "aws_eks_node_group" "eks_node_group" {
  cluster_name           = aws_eks_cluster.kubernetes_cluster.name
  subnet_ids             = [local.main_subnet_id] # Limit nodes to one subnet
  node_role_arn          = aws_iam_role.dss-cluster-node-group.arn
  node_group_name_prefix = aws_eks_cluster.kubernetes_cluster.name
  instance_types = [
    var.aws_instance_type
  ]

  scaling_config {
    desired_size = var.node_count
    max_size     = var.node_count
    min_size     = var.node_count
  }

  lifecycle {
    create_before_destroy = true
  }

  depends_on = [
    aws_eip.gateway,
    aws_eip.ip_crdb
  ]
}
