resource "helm_release" "storage-class" {
  count = var.aws_create_storage_class ? 1 : 0

  chart = "${path.module}/charts/storage-class"
  name  = "storage-class"
  wait  = true
  depends_on = [
    aws_eks_addon.aws-ebs-csi-driver
  ]
}
