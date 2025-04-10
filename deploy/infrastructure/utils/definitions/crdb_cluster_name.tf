variable "crdb_cluster_name" {
  type        = string
  description = <<-EOT
  A string that specifies a CRDB cluster name. This is used together to ensure that all newly created
  nodes join the intended cluster when you are running multiple clusters.
  The CRDB cluster is automatically given a randomly-generated name if an empty string is provided.
  The CRDB cluster name must be 6-20 characters in length, and can include lowercase letters, numbers,
  and dashes (but no leading or trailing dashes). A cluster's name cannot be edited after it is created.

  At the moment, this variable is only used for helm charts deployments.

  Example: interuss_us_production
  EOT
}
