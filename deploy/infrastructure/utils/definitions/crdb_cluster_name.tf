variable "crdb_cluster_name" {
  type        = string
  description = <<-EOT
  A string that specifies a CRDB cluster name. This is used together to ensure that all newly created
  nodes join the intended cluster when you are running multiple clusters.
  The CRDB cluster is automatically given a randomly-generated name if an empty string is provided.
  The CRDB cluster name must be 1-256 characters in length, start with a letter, and can include letters,
  numbers, and dashes (but no trailing dash). A cluster's name cannot be edited after it is created.

  Example: interuss-us-production
  EOT
}
