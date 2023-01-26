resource "aws_eks_cluster" "kubernetes_cluster" {
  name     = var.cluster_name
  role_arn = aws_iam_role.dss-cluster.arn


  vpc_config {
    subnet_ids = aws_subnet.dss[*].id
    endpoint_public_access = true
    public_access_cidrs = [
      "0.0.0.0/0"
    ]
  }

  # Ensure that IAM Role permissions are created before and deleted after EKS Cluster handling.
  # Otherwise, EKS will not be able to properly delete EKS managed EC2 infrastructure such as Security Groups.
  depends_on = [
    aws_iam_role_policy_attachment.dss-cluster-service,
    aws_internet_gateway.dss
  ]

  version = "1.24"

#  enabled_cluster_log_types = ["api", "audit", "authenticator", "controllerManager", "scheduler"]
}

resource "aws_eks_node_group" "eks_node_group" {
  cluster_name = aws_eks_cluster.kubernetes_cluster.name
  subnet_ids = aws_subnet.dss[*].id
  node_role_arn = aws_iam_role.dss-cluster-node-group.arn
  disk_size = 100
  node_group_name_prefix = aws_eks_cluster.kubernetes_cluster.name
  instance_types = [
    var.aws_instance_type
  ]

  scaling_config {
    desired_size = 2
    max_size     = 3
    min_size     = 1
  }

  remote_access {
    ec2_ssh_key = "test-1"
  }
}
