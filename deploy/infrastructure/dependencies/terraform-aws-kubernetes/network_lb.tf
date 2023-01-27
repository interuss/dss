
# Load Balancer Kubernetes Controller
resource helm_release "aws-load-balancer-controller" {
  repository = "https://aws.github.io/eks-charts"
  chart = "aws-load-balancer-controller"
  name = "aws-load-balancer-controller"

  namespace = "kube-system"

  set {
    name  = "clusterName"
    value = var.cluster_name
  }

  depends_on = [
    aws_eks_cluster.kubernetes_cluster
  ]
}

# SSL Certificate
resource aws_acm_certificate "app_hostname" {
  domain_name       = var.app_hostname
  validation_method = "DNS"

  lifecycle {
    create_before_destroy = true
  }
}

## DNS records for SSL Certificate validation
resource "aws_route53_record" "app_hostname_cert_validation" {
  count = 1

  allow_overwrite = true
  name            = element(aws_acm_certificate.app_hostname.domain_validation_options.*.resource_record_name, count.index)
  type            = element(aws_acm_certificate.app_hostname.domain_validation_options.*.resource_record_type, count.index)
  records         = [element(aws_acm_certificate.app_hostname.domain_validation_options.*.resource_record_value, count.index)]
  ttl             = 60
  zone_id         = var.aws_route53_zone_id
}

resource "aws_acm_certificate_validation" "app_hostname_cert" {
  certificate_arn         = aws_acm_certificate.app_hostname.arn
  validation_record_fqdns = aws_route53_record.app_hostname_cert_validation[*].fqdn
}

output "app_hostname_cert_arn" {
  value = aws_acm_certificate.app_hostname.arn
}

# Public Elastic IP for the gateway (1 per subnet)
# At the moment, worker nodes will be deployed in the same subnet, so only one elastic ip is required.
resource aws_eip "gateway" {
  vpc = true
  count = 1
}

# Public Elastic IPs for the crdb instances
resource aws_eip "ip_crdb" {
  count  = var.node_count
  vpc = true

  tags = {
    Name = format("%s-ip-crdb%v", var.cluster_name, count.index)
    # Preserve mapping between ips and hostnames
    ExpectedDNS = format("%s.%s", count.index, var.crdb_hostname_suffix)
  }
}

# Application DNS
resource aws_route53_record "app_hostname" {
  count = var.aws_route53_zone_id == null ? 0 : 1
  zone_id = var.aws_route53_zone_id
  name    = var.app_hostname
  type    = "A"
  ttl     = 300
  records = aws_eip.gateway[*].public_ip
}
