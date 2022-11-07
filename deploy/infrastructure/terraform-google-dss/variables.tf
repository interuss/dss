variable "google_cluster_context" {
  type = object({
    # Name of the new cluster.
    name = string

    # Name of the GCP project hosting the future cluster.
    region = string

    # GCP Region where to deploy the cluster.
    zone = string

    # GCP Zone where to deploy the cluster
    project = string

    # GCP machine type used for the Kubernetes node pool.
    # Example: n2-standard-4 for production, e2-micro for development
    dns_managed_zone_name = optional(string, "")

    # GCP DNS zone name to automatically manage DNS entries. Leave it empty to manage it manually.
    machine_type = optional(string, "n2-standard-4")
  })

  validation {
    condition     = startswith(var.google_cluster_context.zone, var.google_cluster_context.region)
    error_message = "The zone shall be part of the region"
  }
}

# Intentionally left undefined to reuse schema validation from terraform-commons-dss
variable "dss_configuration" {}