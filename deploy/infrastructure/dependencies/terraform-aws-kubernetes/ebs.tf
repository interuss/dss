data "tls_certificate" "cluster_oidc_provider" {
  url = aws_eks_cluster.kubernetes_cluster.identity[0].oidc[0].issuer
}

resource aws_iam_openid_connect_provider "cluster_provider" {
  client_id_list = ["sts.amazonaws.com"]
  thumbprint_list = data.tls_certificate.cluster_oidc_provider.certificates[*].sha1_fingerprint
  url             = data.tls_certificate.cluster_oidc_provider.url
}


resource "local_file" "setup-eks-service-accounts" {
  content  = templatefile("${path.module}/templates/setup-eks-service-accounts.sh.tmp", {
    ebs_csi_controller_sa_role_arn = aws_iam_role.AmazonEKS_EBS_CSI_DriverRole.arn
  })
  filename = "${path.cwd}/setup-eks-service-accounts.sh"
}