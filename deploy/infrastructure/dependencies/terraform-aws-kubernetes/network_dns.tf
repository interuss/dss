
locals {
  crdb_hostnames = var.aws_route53_zone_id == "" ? {} : { for i in aws_eip.ip_crdb[*] : i.tags.ExpectedDNS => i.public_ip }
  yugabyte_master_hostnames = var.aws_route53_zone_id == "" ? {} : { for i in aws_eip.ip_yugabyte_masters[*] : i.tags.ExpectedDNS => i.public_ip }
  yugabyte_tserver_hostnames = var.aws_route53_zone_id == "" ? {} : { for i in aws_eip.ip_yugabyte_tservers[*] : i.tags.ExpectedDNS => i.public_ip }
}


## DNS records for SSL Certificate validation
resource "aws_route53_record" "app_hostname_cert_validation" {
  count = var.aws_route53_zone_id == "" ? 0 : length(aws_acm_certificate.app_hostname.domain_validation_options)

  allow_overwrite = true
  name            = element(aws_acm_certificate.app_hostname.domain_validation_options.*.resource_record_name, count.index)
  type            = element(aws_acm_certificate.app_hostname.domain_validation_options.*.resource_record_type, count.index)
  records         = [element(aws_acm_certificate.app_hostname.domain_validation_options.*.resource_record_value, count.index)]
  ttl             = 60
  zone_id         = var.aws_route53_zone_id
}

# Application DNS
resource "aws_route53_record" "app_hostname" {
  count = var.aws_route53_zone_id == "" ? 0 : length(aws_eip.gateway)

  zone_id = var.aws_route53_zone_id
  name    = var.app_hostname
  type    = "A"
  ttl     = 300
  records = [aws_eip.gateway[count.index].public_ip]
}

# Crdb nodes DNS
resource "aws_route53_record" "crdb_hostname" {
  for_each = local.crdb_hostnames

  zone_id = var.aws_route53_zone_id
  name    = each.key
  type    = "A"
  ttl     = 300
  records = [each.value]
}

# Yugabyte master nodes DNS
resource "aws_route53_record" "yugabyte_master_hostnames" {
  for_each = local.yugabyte_master_hostnames

  zone_id = var.aws_route53_zone_id
  name    = each.key
  type    = "A"
  ttl     = 300
  records = [each.value]
}

# Yugabyte tserver nodes DNS
resource "aws_route53_record" "yugabyte_tserver_hostnames" {
  for_each = local.yugabyte_tserver_hostnames

  zone_id = var.aws_route53_zone_id
  name    = each.key
  type    = "A"
  ttl     = 300
  records = [each.value]
}
