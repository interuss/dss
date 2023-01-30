# This file is an example, please adapt it to your configuration.
# See TFVARS.md for full set of variables and related descriptions.

# AWS account
aws_region = "eu-west-1"

# DNS
aws_route53_zone_id  = "Z01551234567890123456"
app_hostname         = "dss.interuss.example.com"
crdb_hostname_suffix = "db.interuss.example.com"

# Kubernetes configuration
cluster_name                 = "dss-dev-ew1"
node_count                   = 3
aws_instance_type            = "t3.medium"
aws_kubernetes_storage_class = "gp2"

# DSS configuration
image = "latest"
authorization = {
  public_key_pem_path = "/test-certs/auth2.pem"
}
should_init         = true
crdb_locality       = "interuss_dss-aws-ew1"
crdb_external_nodes = []