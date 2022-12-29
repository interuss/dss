# Variables used to interface with other modules of type terraform-*-kubernetes.

variable "kubernetes_cloud_provider_name" {
  type        = string
  description = "Cloud provider name"
}

variable "kubernetes_get_credentials_cmd" {
  type        = string
  description = "Command to get credentials to access the Kubernetes cluster"
}

variable "kubernetes_context_name" {
  type        = string
  description = "Cluster context name used by kubectl to access the Kubernetes cluster"
}

variable "kubernetes_api_endpoint" {
  type        = string
  description = "Kubernetes cluster API endpoint"
}

# Hostnames and DNS
variable "crdb_internal_nodes" {
  type = list(object({
    dns = string
    ip  = string
  }))
  description = "List of the IP addresses and related dns for the Cockroach DB nodes"
}

variable "ip_gateway" {
  type = string
  description = "IP of the gateway used by the DSS service"
}