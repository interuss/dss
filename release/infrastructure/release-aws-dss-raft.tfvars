# AWS account
aws_region = "us-east-2"

# DNS Management
aws_route53_zone_id = "$ZONE_ID"

# Hostnames
app_hostname       = "dss.release-raft.ci.aws-interuss.uspace.dev" # Remark: All endpoints are public and not a secret
db_hostname_suffix = "db.release-raft.ci.aws-interuss.uspace.dev"

# Kubernetes configuration
cluster_name                 = "dss-r-raft"
kubernetes_version           = 1.36
node_count                   = 3
aws_instance_type            = "m5.xlarge"
aws_kubernetes_storage_class = "gp3"

# DSS configuration
image = "$IMAGE"
authorization = {
  public_key_pem_path = "/test-certs/auth2.pem"
}
should_init = true

# CockroachDB
crdb_image_tag      = "v24.1.3"
crdb_cluster_name   = "interuss-example"
locality            = "interuss_dss-aws-ue2"
crdb_external_nodes = []

datastore_type = "raft"

raft_automatic_bootstrap       = true
raft_automatic_bootstrap_nodes = ["0.db.release-raft.ci.google-interuss.uspace.dev:2407", "1.db.release-raft.ci.google-interuss.uspace.dev:2407", "2.db.release-raft.ci.google-interuss.uspace.dev:2407"]

prometheus_hostname = "prometheus.release-raft.ci.aws-interuss.uspace.dev"
enable_monitoring   = true
