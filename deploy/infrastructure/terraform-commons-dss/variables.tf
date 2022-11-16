# Public schema used as input by provider specific modules.
variable "dss_configuration" {
  type = object({
    # Kubernetes

    # Namespace where to deploy Kubernetes resources
    # TODO: Adapt current deployment scripts to support default is supported for the moment.
    namespace = optional(string, "default")

    # Full name of the docker image built in the section above. build.sh prints this name as
    # the last thing it does when run with DOCKER_URL set. It should look something like
    # gcr.io/your-project-id/dss:2020-07-01-46cae72cf if you built the image yourself, or docker.io/interuss/dss
    # TODO: Test DOCKER_URL usage as documented in /build/README.md
    image = optional(string, "docker.io/interuss/dss:v0.4.0")

    # Infrastructure

    # Number of pool nodes.
    # Example: 3 for production. 1 for development.
    crdb_node_count = optional(number, 3)

    # Kubernetes Storage Class to use for CockroachDB and Prometheus volumes. You can
    # check your cluster's possible values with kubectl get storageclass.
    # Example standard
    storage_class = string


    # DSS Functionalities

    # Set this boolean true to enable ASTM strategic conflict detection functionality.
    enable_scd = bool

    # Set to false if joining an existing pool, true if creating the first DSS instance
    # for a pool. When set true, this can initialize the data directories on your cluster,
    # and prevent you from joining an existing pool.
    should_init = optional(bool, false)

    # Fully-qualified domain name of your HTTPS Gateway ingress endpoint.
    # Example, dss.example.com.
    app_hostname = string


    # Authorization

    # If providing a .pem file directly as the public key to validate incoming access tokens, specify the name
    # of this .pem file here as /public-certs/YOUR-KEY-NAME.pem replacing YOUR-KEY-NAME as appropriate. For instance,
    # if using the provided us-demo.pem, use the path /public-certs/us-demo.pem. Note that your .pem file should built
    # in the docker image or mounted manually.
    # TODO: Add ability to provide the key content
    public_key_pem_path = string

    # If providing the access token public key via JWKS, specify the JWKS endpoint here.
    # Example: https://auth.example.com/.well-known/jwks.json
    jwks_endpoint = string

    # If providing the access token public key via JWKS, specify the kid (key ID) of they appropriate key in the JWKS file referenced above.
    # If providing a .pem file directly as the public key to valid incoming access tokens, provide a blank string for this parameter.
    jwks_key_id = string

    # Database

    # The domain name suffix shared by all of your CockroachDB nodes. For instance,
    # if your CRDB nodes were addressable at 0.db.example.com, 1.db.example.com, and
    # 2.db.example.com, then VAR_CRDB_HOSTNAME_SUFFIX would be db.example.com.
    # Example: db.example.com
    crdb_hostname_suffix = string

    # Unique name for your DSS instance. Currently, we recommend "<ORG_NAME>_<CLUSTER_NAME>",
    # and the = character is not allowed. However, any unique (among all other participating
    # DSS instances) value is acceptable.
    crdb_locality = string # <ORGNAME_CLUSTER_NAME>

    # Fully-qualified domain name of your HTTPS Gateway ingress endpoint. For example, dss.example.com.
    crdb_external_nodes = list(string)

    # Desired schemas versions
    desired_rid_db_version = optional(string, "4.0.0")
    desired_scd_db_version = optional(string, "3.1.0")
  })

  validation {
    condition     = var.dss_configuration.should_init == true && length(var.dss_configuration.crdb_external_nodes) == 0
    error_message = "crdb_external_nodes should be empty when should_init is set to true"
  }

  # TODO: Adapt current deployment scripts in /build/deploy to support default is supported for the moment.
  validation {
    condition     = var.dss_configuration.namespace == "default"
    error_message = "Only default namespace is supported at the moment"
  }
}

# Internal schema
variable "kubernetes" {
  type = object({
    provider_name                = string
    get_credentials_cmd          = string
    kubectl_cluster_context_name = string
    api_endpoint                 = string
    node_addresses               = list(string)
    crdb_nodes = list(object({
      dns = string
      ip  = string
    }))
    ip_gateway = string
  })
}