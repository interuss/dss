# This module is expected to be applied by the Github CI user. By default, only the user has permission to
# connect to the cluster. This file gathers resources to grant access to AWS administrators.

resource "kubernetes_config_map_v1_data" "aws-auth" {
  metadata {
    name      = "aws-auth"
    namespace = "kube-system"
  }

  force = true # EKS provisions this file by default.

  data = {
    mapRoles = yamlencode([
      {
        groups   = [
          "system:bootstrappers",
          "system:nodes"
        ]
        rolearn  = module.terraform-aws-kubernetes.iam_role_node_group_arn
        username = "system:node:{{EC2PrivateDNSName}}"
      },
      {
        groups   = [
          "system:masters"
        ]
        rolearn  = var.aws_iam_administrator_role
        username = "aws-administrator"
      }
    ])
  }
}
