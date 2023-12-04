aws_region = "eu-west-1"
cluster_name = "dss-1"
aws_route53_zone_id = "Z01554482LNPDVY7FW95I"
app_hostname = "dss.aws-interuss.uspace.dev"
aws_instance_type = "t3.medium"
crdb_hostname_suffix = "db.aws-interuss.uspace.dev"
node_count = 3

aws_kubernetes_storage_class = "gp2"

# DSS configuration
image = "latest"
authorization = {
  public_key_pem_path = "/test-certs/auth2.pem"
}
should_init   = true
crdb_locality = "interuss_dss-aws-ew1"
crdb_external_nodes = []
