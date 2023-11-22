# This file is an example, please adapt it to your configuration.
# See TFVARS.md for the full set of variables and related descriptions.

# AWS account
aws_region = "eu-west-1"

# DNS Management
aws_route53_zone_id = ""

# Hostnames
app_hostname = "dss.aws-interuss-ci.uspace.dev"
crdb_hostname_suffix = "db.aws-interuss-ci.uspace.dev"

# Kubernetes configuration
cluster_name                 = "dss-ci-aws-ew1"
node_count                   = 3
aws_instance_type            = "t3.medium"
aws_kubernetes_storage_class = "gp2"

# DSS configuration
image = "latest"
authorization = {
  public_key_pem_path = "/test-certs/auth2.pem"
}
should_init         = true
crdb_locality       = "interuss_dss-ci-aws-ew1"
crdb_external_nodes = []