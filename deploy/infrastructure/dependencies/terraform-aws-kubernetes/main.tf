terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">=4.0"
    }
    tls = {
      source = "hashicorp/tls"
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

provider "helm" {
  kubernetes {
    host                   = aws_eks_cluster.kubernetes_cluster.endpoint
    cluster_ca_certificate = base64decode(aws_eks_cluster.kubernetes_cluster.certificate_authority[0].data)
    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      args        = ["eks", "get-token", "--cluster-name", var.cluster_name]
      command     = "aws"
    }
  }
}