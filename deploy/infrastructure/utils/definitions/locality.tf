variable "locality" {
  type        = string
  description = <<-EOT
  Unique name for your DSS instance, in the form "zone=<ID>". Currently, we recommend
  "zone=<ORG_NAME>_<CLUSTER_NAME>". However, any unique (among all other participating
  DSS instances) <ID> is acceptable.
  With CockroachDB, this value is passed as is to the --locality flag of the nodes.

  Example: zone=<ORGNAME_CLUSTER_NAME>
  EOT

  validation {
    condition     = var.locality != ""
    error_message = "Locality value must be set"
  }
}
