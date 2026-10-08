resource "aws_eks_cluster" "kubernetes_cluster" {
  name     = var.cluster_name
  role_arn = aws_iam_role.dss-cluster.arn

  vpc_config {
    subnet_ids              = local.vpc_id == "" ? aws_subnet.dss[*].id : var.private_subnet_ids
    endpoint_private_access = true
    endpoint_public_access  = var.use_public_subnets
    public_access_cidrs     = var.use_public_subnets ? local.public_access_cidrs : null
    security_group_ids      = [aws_security_group.eks-controlplane.id]
  }

  lifecycle {
    precondition {
      condition     = var.vpc_id != "" || var.use_public_subnets
      error_message = "A private-only API endpoint on a newly created VPC is unreachable from the Terraform runner. Supply vpc_id with VPN or in-VPC access, or set use_public_subnets = true."
    }

    precondition {
      condition     = var.vpc_id == "" || (length(var.private_subnet_ids) >= 2 && length(var.public_subnet_ids) >= 1)
      error_message = "When vpc_id is set, provide at least two private subnets (EKS requires two AZs) and at least one public subnet."
    }
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
