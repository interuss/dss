# AWS account
aws_region = "eu-central-1"

# DNS Management
aws_route53_zone_id = "${ZONE_ID}"

# Hostnames
app_hostname       = "dss.release-crdb.ci.aws-interuss.uspace.dev"  # Remark: All endpoints are public and not a secret
db_hostname_suffix = "db.release-crdb.ci.aws-interuss.uspace.dev"

# Kubernetes configuration
cluster_name                 = "dss-r-crdb"
kubernetes_version           = 1.35
node_count                   = 3
aws_instance_type            = "m5.xlarge"
aws_kubernetes_storage_class = "gp2"

# DSS configuration
image = "${IMAGE}"
authorization = {
  public_key_pem_path = "/test-certs/auth2.pem"
}
should_init = true

# CockroachDB
crdb_image_tag      = "v24.1.3"
crdb_cluster_name   = "interuss-example"
locality            = "interuss_dss-aws-ew1"
crdb_external_nodes = []

datastore_type = "cockroachdb"

evict_enable_scd_cron     = true
evict_scd_schedule = "*/2 * * * *"
evict_rid_schedule = "*/2 * * * *"

prometheus_hostname = "prometheus.release-crdb.ci.aws-interuss.uspace.dev"
enable_monitoring   = true
