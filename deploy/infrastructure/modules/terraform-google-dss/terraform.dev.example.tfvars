# This file is an example, please adapt it to your configuration.
# See TFVARS.md for full set of variables and related descriptions.

# Google account
google_project_name = "interuss-deploy-example"
google_zone         = "europe-west6-a"


# DNS
google_dns_managed_zone_name = "interuss-example-com"
app_hostname                 = "dss.interuss.example.com"
crdb_hostname_suffix         = "db.interuss.example.com"

# Kubernetes configuration
cluster_name                    = "dss-dev-w6a"
kubernetes_version              = 1.32
node_count                      = 3
google_machine_type             = "e2-medium"
google_kubernetes_storage_class = "standard"

# DSS configuration
image             = "docker.io/interuss/dss:latest"
image_pull_secret = ""
authorization = {
  public_key_pem_path = "/test-certs/auth2.pem"
}
should_init = true

# Datastore
datastore_type = "cockroachdb"

# CockroachDB
crdb_image_tag      = "v24.1.3"
crdb_cluster_name   = "interuss-example"
crdb_locality       = "interuss_dss-dev-w6a"
crdb_external_nodes = []

# Yugabyte
yugabyte_region = "uss-1"
