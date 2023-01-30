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

variable "kubernetes_storage_class" {
  type        = string
  description = <<-EOT
  Kubernetes Storage Class to use for CockroachDB and Prometheus volumes. You can
  check your cluster's possible values with `kubectl get storageclass`.
  This value is provided by the cloud provider specific variable `*_kubernetes_storage_class`.

  Example: value of `var.google_kubernetes_storage_class`
  EOT
}

variable "gateway_cert_name" {
  type = string
  description = "Only required for AWS cloud provider. Certificate reference used by the DSS Gateway. For AWS, provide the ARN of the certificate."
  default = ""
}

variable "subnet" {
  type = string
  description = "Only required for AWS cloud provider. Subnet where the kubernetes worker nodes is deployed. For AWS, provide the name or the id of the subnet"
  default = ""
}