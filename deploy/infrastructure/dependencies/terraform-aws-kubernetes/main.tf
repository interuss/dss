terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
    tls = {
      source = "hashicorp/tls"
      version = "~> 4.0"
    }
    helm = {
      source = "hashicorp/helm"
      version = "~> 3.0"
    }
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Cluster = var.cluster_name
    }
  }
}

data "aws_eks_cluster_auth" "kubernetes_cluster" {
  name = aws_eks_cluster.kubernetes_cluster.name
}

provider "helm" {
  kubernetes = {
    host                   = aws_eks_cluster.kubernetes_cluster.endpoint
    cluster_ca_certificate = base64decode(aws_eks_cluster.kubernetes_cluster.certificate_authority[0].data)
    token                  = data.aws_eks_cluster_auth.kubernetes_cluster.token
  }
}
