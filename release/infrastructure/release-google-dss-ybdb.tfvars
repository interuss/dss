# Google account
google_project_name = "${GOOGLE_PROJECT_NAME}"
google_zone         = "europe-west3-a"
google_delete_protection = false


# DNS
google_dns_managed_zone_name = "${ZONE_NAME}"
app_hostname                 = "dss.release-ybdb.ci.google-interuss.uspace.dev"  # Remark: All endpoints are public and not a secret
db_hostname_suffix           = "db.release-ybdb.ci.google-interuss.uspace.dev"

# Kubernetes configuration
cluster_name       = "dss-r-ybdb"
kubernetes_version = 1.35
node_count         = 3
google_machine_type             = "n2-standard-4"
google_kubernetes_storage_class = "standard"

# DSS configuration
image = "${IMAGE}"
image_pull_secret = ""
authorization = {
  public_key_pem_path = "/test-certs/auth2.pem"
}
should_init = false

# CockroachDB
crdb_image_tag      = "v24.1.3"
crdb_cluster_name   = "interuss-example"
locality            = "interuss_dss-dev-w6a"
crdb_external_nodes = []

datastore_type = "yugabyte"

yugabyte_region = "gke-uss-1"
yugabyte_zone   = "gke-uss-1"
yugabyte_light_resources = false
yugabyte_external_nodes = ["0.master.db.release-ybdb.ci.aws-interuss.uspace.dev", "1.master.db.release-ybdb.ci.aws-interuss.uspace.dev", "2.master.db.release-ybdb.ci.aws-interuss.uspace.dev"]

evict_enable_scd_cron     = true
evict_scd_schedule = "*/2 * * * *"
evict_rid_schedule = "*/2 * * * *"

prometheus_hostname = "prometheus.release-ybdb.ci.google-interuss.uspace.dev"
enable_monitoring   = true
