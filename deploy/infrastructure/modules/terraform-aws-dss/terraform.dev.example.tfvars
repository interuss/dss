# This file is an example, please adapt it to your configuration.
# See TFVARS.md for the full set of variables and related descriptions.

# AWS account
aws_region = "eu-west-1"

# DNS Management
aws_route53_zone_id = "Z01551234567890123456"

# VPC configuration. For securirty, it is recommended to use private subnets for production environments.
# For backward compatibility, set use_public_subnets true, vpc_id to a blank string and subnets to empty lists
use_public_subnets = false
vpc_id = "vpc-01234567890123456"
private_subnet_ids = ["subnet-01234567890123456", "subnet-01234567890123456"]
public_subnet_ids = ["subnet-01234567890123456", "subnet-01234567890123456"]

# Hostnames
app_hostname       = "dss.interuss.example.com"
db_hostname_suffix = "db.interuss.example.com"

# Kubernetes configuration
cluster_name                 = "dss-dev-ew1"
kubernetes_version           = 1.32
node_count                   = 3
aws_instance_type            = "t3.medium"
aws_kubernetes_storage_class = "gp3"

# DSS configuration
image = "docker.io/interuss/dss:latest"
authorization = {
  public_key_pem_path = "/test-certs/auth2.pem"
}

# Datastore
datastore_type = "cockroachdb"

# CockroachDB
crdb_image_tag      = "v24.1.3"
crdb_cluster_name   = "interuss_example"
locality            = "interuss_dss-aws-ew1"
crdb_external_nodes = []
should_init         = true

# Yugabyte
yugabyte_region          = "aws-uss-1"
yugabyte_zone            = "aws-uss-1"
yugabyte_light_resources = false
yugabyte_external_nodes  = []

# Monitoring
enable_monitoring = false
# prometheus_hostname             = "prometheus.dss.interuss.example.com"
