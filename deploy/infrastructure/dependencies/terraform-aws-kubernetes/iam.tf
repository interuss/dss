data "aws_caller_identity" "current" {}

locals {
  aws_account_id          = data.aws_caller_identity.current.account_id
  aws_cluster_id          = aws_eks_cluster.kubernetes_cluster.id
  aws_cluster_oidc_issuer = aws_eks_cluster.kubernetes_cluster.identity[0].oidc[0].issuer
}

resource "aws_iam_role" "dss-cluster" {
  name = "${var.cluster_name}-dss-cluster"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "eks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

# Policy used by internal kubernetes services to access AWS resources.
resource "aws_iam_role_policy_attachment" "dss-cluster-service" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.dss-cluster.name
}

resource "aws_iam_role" "dss-cluster-node-group" {
  name = "${var.cluster_name}-cluster-node-group"

  assume_role_policy = jsonencode({
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
    Version = "2012-10-17"
  })
}

resource "aws_iam_policy" "AWSLoadBalancerControllerPolicy" {
  name = "${var.cluster_name}-AWSLoadBalancerControllerPolicy"
  # Source: https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html
  # Template: https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/v2.4.4/docs/install/iam_policy.json
  policy = file("${path.module}/AWSLoadBalancerControllerPolicy.json")
}

resource "aws_iam_role_policy_attachment" "AWSLoadBalancerControllerPolicy" {
  policy_arn = aws_iam_policy.AWSLoadBalancerControllerPolicy.arn
  role       = aws_iam_role.dss-cluster-node-group.name
}

resource "aws_iam_role_policy_attachment" "ElasticLoadBalancingFullAccess" {
  policy_arn = "arn:aws:iam::aws:policy/ElasticLoadBalancingFullAccess"
  role       = aws_iam_role.dss-cluster-node-group.name
}

resource "aws_iam_role_policy_attachment" "AmazonEKSWorkerNodePolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = aws_iam_role.dss-cluster-node-group.name
}

resource "aws_iam_role_policy_attachment" "AmazonEKS_CNI_Policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.dss-cluster-node-group.name
}

## Docker registry
resource "aws_iam_role_policy_attachment" "AmazonEC2ContainerRegistryReadOnly" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.dss-cluster-node-group.name
}

## EBS
resource "aws_iam_role" "AmazonEKS_EBS_CSI_DriverRole" {
  name = "${var.cluster_name}-AmazonEKS_EBS_CSI_DriverRole"

  assume_role_policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Effect" : "Allow",
        "Principal" : {
          "Federated" : format("arn:aws:iam::${local.aws_account_id}:%s", replace(local.aws_cluster_oidc_issuer, "https://", "oidc-provider/")),
        },
        "Action" : "sts:AssumeRoleWithWebIdentity",
        "Condition" : {
          "StringEquals" : {
            format("%s:aud", replace(local.aws_cluster_oidc_issuer, "https://", "")) : "sts.amazonaws.com",
            format("%s:sub", replace(local.aws_cluster_oidc_issuer, "https://", "")) : "system:serviceaccount:kube-system:ebs-csi-controller-sa"
          }
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "AmazonEKS_EBS_CSI_DriverRole" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy"
  role       = aws_iam_role.AmazonEKS_EBS_CSI_DriverRole.name
}
