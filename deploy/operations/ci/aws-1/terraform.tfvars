# See TFVARS.md for the full set of variables and related descriptions.

# AWS account
aws_region = "us-east-1"

# DNS Management
aws_route53_zone_id = "Z03377073HUSGB4L9FKEK"

# Hostnames
app_hostname       = "dss.ci.aws-interuss.uspace.dev"
db_hostname_suffix = "db.ci.aws-interuss.uspace.dev"

# Kubernetes configuration
kubernetes_version           = 1.32
cluster_name                 = "dss-ci-aws-ue1"
node_count                   = 3
aws_instance_type            = "t3.medium"
aws_kubernetes_storage_class = "gp2"

# DSS configuration
image = "docker.io/interuss/dss:latest"
authorization = {
  public_key_pem_path = "/test-certs/auth2.pem"
}
should_init         = true
crdb_image_tag      = "v24.1.3"
crdb_cluster_name   = "interuss-ci"
locality            = "interuss_dss-ci-aws-ue1"
crdb_external_nodes = []

aws_iam_permissions_boundary = "arn:aws:iam::301042233698:policy/GithubCIPermissionBoundaries20231130225039606500000001"
aws_iam_administrator_role   = "arn:aws:iam::301042233698:role/AWSReservedSSO_AdministratorAccess_9b637c80b830ea2c"
aws_iam_ci_role              = "arn:aws:iam::301042233698:role/InterUSSGithubCI"
