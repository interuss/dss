// This module is expected to be applied by the Github CI user. By default, only the user who created the cluster
// has permission to connect to the cluster. This file gathers resources to grant access to AWS administrators.

resource "local_file" "aws-auth-config-map" {
  content = yamlencode({
    apiVersion = "v1"
    kind = "ConfigMap"
    metadata = {
      name      = "aws-auth"
      namespace = "kube-system"
    }
    data = {
      mapRoles = yamlencode([
        {
          groups   = [
            "system:bootstrappers",
            "system:nodes"
          ]
          rolearn  = module.terraform-aws-dss.iam_role_node_group_arn
          username = "system:node:{{EC2PrivateDNSName}}"
        },
        {
          groups   = [
            "system:masters"
          ]
          rolearn  = var.aws_iam_administrator_role
          username = "interuss-aws-administrator"
        },
        {
          groups   = [
            "system:masters"
          ]
          rolearn  = var.aws_iam_ci_role
          username = "interuss-ci"
        }
      ])
    }
  })

  filename   = "${module.terraform-aws-dss.workspace_location}/aws_auth_config_map.yml"
}
