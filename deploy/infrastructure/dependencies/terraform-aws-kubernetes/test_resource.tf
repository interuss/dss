
resource "local_file" "test-app" {
  filename = "test-app.yml"
  content = templatefile("${path.module}/test-app.template.yml", {
    certificate_arn = aws_acm_certificate.app_hostname.arn
    eip_alloc_ids = aws_eip.gateway[*].allocation_id
    loadbalancer_name = "${var.cluster_name}-lb"
    subnet_ids = [aws_subnet.dss[0].id]
  })
}