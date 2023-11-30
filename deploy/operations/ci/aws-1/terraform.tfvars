# See TFVARS.md for the full set of variables and related descriptions.

# AWS account
aws_region = "us-east-1"

# DNS Management
aws_route53_zone_id = "Z03377073HUSGB4L9FKEK"

# Hostnames
app_hostname = "dss.ci.aws-interuss.uspace.dev"
crdb_hostname_suffix = "db.ci.aws-interuss.uspace.dev"

# Kubernetes configuration
cluster_name                 = "dss-ci-aws-ue1"
node_count                   = 3
aws_instance_type            = "t3.medium"
aws_kubernetes_storage_class = "gp2"

# DSS configuration
image = "latest"
authorization = {
  public_key_pem_path = "/test-certs/auth2.pem"
}
should_init         = true
crdb_locality       = "interuss_dss-ci-aws-ue1"
crdb_external_nodes = []