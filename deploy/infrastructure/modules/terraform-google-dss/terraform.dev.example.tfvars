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
node_count                      = 3
google_machine_type             = "e2-medium"
google_kubernetes_storage_class = "standard"

# DSS configuration
image = "latest"
authorization = {
  public_key_pem_path = "/jwt-public-certs/us-demo.pem"
}
should_init   = true
crdb_locality = "interuss_dss-dev-w6a"

crdb_external_nodes = []