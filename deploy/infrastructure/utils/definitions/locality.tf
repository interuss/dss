variable "locality" {
  type        = string
  description = <<-EOT
  Unique name for your DSS instance. Currently, we recommend "<ORG_NAME>_<CLUSTER_NAME>",
  and the = character is not allowed. However, any unique (among all other participating
  DSS instances) value is acceptable.

  Example: <ORGNAME_CLUSTER_NAME>
  EOT

  validation {
    condition     = var.locality != ""
    error_message = "Locality value must be set"
  }
}
