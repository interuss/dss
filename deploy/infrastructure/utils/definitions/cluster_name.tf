variable "cluster_name" {
  type = string
  description = <<-EOT
    Name of the kubernetes cluster that will host this DSS instance (should generally describe the DSS instance being hosted)

    Example: `dss-che-1`
  EOT
}