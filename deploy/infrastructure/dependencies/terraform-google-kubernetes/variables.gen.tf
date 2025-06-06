
# This file has been automatically generated by /deploy/infrastructure/utils/generate_terraform_variables.py.
# Please do not modify manually.

variable "google_project_name" {
  type        = string
  description = "Name of the GCP project hosting the future cluster"
}

variable "google_zone" {
  type        = string
  description = <<-EOT
  GCP zone hosting the kubernetes cluster
  List of available zones: https://cloud.google.com/compute/docs/regions-zones#available

  Example: `europe-west6-a`
  EOT
}

variable "google_dns_managed_zone_name" {
  type        = string
  description = <<-EOT
  GCP DNS zone name to automatically manage DNS entries.

  Leave it empty to manage it manually.
  EOT
}

variable "google_machine_type" {
  type        = string
  description = <<-EOT
  GCP machine type used for the Kubernetes node pool.
  Example: `n2-standard-4` for production, `e2-medium` for development
  EOT
}

variable "app_hostname" {
  type        = string
  description = <<-EOT
  Fully-qualified domain name of your HTTPS Gateway ingress endpoint.

  Example: `dss.example.com`
  EOT
}

variable "crdb_hostname_suffix" {
  type        = string
  description = <<-EOT
  The domain name suffix shared by all of your CockroachDB nodes.
  For instance, if your CRDB nodes were addressable at 0.db.example.com,
  1.db.example.com and 2.db.example.com, then the value would be db.example.com.

  Example: db.example.com
  EOT
}

variable "cluster_name" {
  type        = string
  description = <<-EOT
  Name of the kubernetes cluster that will host this DSS instance (should generally describe the DSS instance being hosted)

  Example: `dss-che-1`
  EOT
}

variable "node_count" {
  type        = number
  description = <<-EOT
  Number of Kubernetes nodes which should correspond to the desired CockroachDB nodes.
  Currently, only single node or three nodes deployments are supported.

  Example: `3`
  EOT

  validation {
    condition     = contains([1, 3], var.node_count)
    error_message = "Currently, only 1 node or 3 nodes deployments are supported."
  }
}


variable "kubernetes_version" {
  type        = string
  description = <<-EOT
  Desired version of the Kubernetes cluster control plane and nodes.

  Supported versions: 1.24 to 1.32
  EOT

  validation {
    condition     = contains(["1.24", "1.25", "1.26", "1.27", "1.28", "1.29", "1.30", "1.31", "1.32"], var.kubernetes_version)
    error_message = "Supported versions: 1.24 to 1.32"
  }
}


