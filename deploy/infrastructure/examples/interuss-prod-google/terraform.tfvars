google_cluster_context = {
  # Name of the new cluster.
  name = "interuss-mini-w6a"

  # Name of the GCP project hosting the future cluster.
  project = ""

  # GCP Region where to deploy the cluster.
  region = "europe-west6"

  # GCP Zone where to deploy the cluster
  zone = "europe-west6-a"

  # GCP machine type used for the Kubernetes node pool.
  # Example: n2-standard-4 for production, e2-micro for development
  machine_type = "n2-standard-4"

  # GCP DNS zone name to automatically manage DNS entries. Leave it empty to manage it manually.
  dns_managed_zone_name = ""
}

dss_configuration = {
  # See build/README.md (Deploying a DSS via Kubernetes, section 11) for variables description.

  namespace = "default"

  # image = "" # Use default. VAR_DOCKER_IMAGE_NAME

  storage_class = "standard" # VAR_STORAGE_CLASS

  enable_scd = true # VAR_ENABLE_SCD

  should_init = true # VAR_SHOULD_INIT

  app_hostname = "" # VAR_APP_HOSTNAME

  public_key_pem_path = "" # VAR_PUBLIC_KEY_PEM_PATH

  jwks_endpoint = "" # VAR_JWKS_ENDPOINT

  jwks_key_id = "" # VAR_JWKS_KEY_ID

  crdb_hostname_suffix = "interuss.example.com" # VAR_CRDB_HOSTNAME_SUFFIX

  crdb_external_nodes = [] # VAR_EXTERNAL_CRDB_NODEn

  crdb_locality = "" # VAR_CRDB_LOCALITY
}
