# Google account
google_project_name      = "$GOOGLE_PROJECT_NAME"
google_zone              = "us-central1-a"
google_delete_protection = false


# DNS
google_dns_managed_zone_name = "$ZONE_NAME"
app_hostname                 = "dss.release-raft.ci.google-interuss.uspace.dev" # Remark: All endpoints are public and not a secret
db_hostname_suffix           = "db.release-raft.ci.google-interuss.uspace.dev"

# Kubernetes configuration
cluster_name                    = "dss-r-raft"
kubernetes_version              = 1.36
node_count                      = 3
google_machine_type             = "n2-standard-4"
google_kubernetes_storage_class = "standard"

# DSS configuration
image             = "$IMAGE"
image_pull_secret = ""
authorization = {
  public_key_pem_path = "/test-certs/auth2.pem"
}
should_init = false

# CockroachDB
crdb_image_tag      = "v24.1.3"
crdb_cluster_name   = "interuss-example"
locality            = "interuss_dss-dev-uc1a"
crdb_external_nodes = []

datastore_type = "raft"

prometheus_hostname = "prometheus.release-raft.ci.google-interuss.uspace.dev"
enable_monitoring   = true
