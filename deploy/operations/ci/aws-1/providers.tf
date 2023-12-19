provider "aws" {
  region = "us-east-1"
}

data "aws_eks_cluster_auth" "kubernetes_cluster" {
  name = var.cluster_name
  depends_on = [module.terraform-aws-kubernetes]
}

data "aws_eks_cluster" "kubernetes_cluster" {
  name = var.cluster_name
  depends_on = [module.terraform-aws-kubernetes]
}

provider kubernetes {
  host                   = data.aws_eks_cluster.kubernetes_cluster.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.kubernetes_cluster.certificate_authority[0].data)
  token                  = data.aws_eks_cluster_auth.kubernetes_cluster.token
}
