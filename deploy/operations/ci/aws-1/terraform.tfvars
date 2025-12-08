# See TFVARS.md for the full set of variables and related descriptions.

# AWS account
aws_region = "us-east-1"

# Kubernetes configuration
kubernetes_version           = 1.32
cluster_name                 = "dss-ci-aws-ue1"
node_count                   = 3
aws_instance_type            = "t3.medium"
aws_kubernetes_storage_class = "gp2"

# DSS configuration
image               = "docker.io/interuss/dss:latest"
should_init         = true
crdb_image_tag      = "v24.1.3"
crdb_cluster_name   = "interuss-ci"
locality            = "interuss_dss-ci-aws-ue1"
crdb_external_nodes = []

# The following variables are injected using the TF_VAR_name method.
# Reference: https://developer.hashicorp.com/terraform/cli/config/environment-variables#tf_var_name

# DNS
# aws_route53_zone_id

# Hostnames
# app_hostname
# db_hostname_suffix

# DSS configuration
# authorization # Value to be provided as json

# AWS
# aws_iam_permissions_boundary
# aws_iam_administrator_role
# aws_iam_ci_role
