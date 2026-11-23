variable "certificates_organization" {
  type        = string
  description = <<-EOT
  Name of the organization operating this DSS instance, passed as --organization to dss-certs.py by the
  generated dss-certs.sh helper. The combination of the kubernetes context name and this organization
  shall be unique in the pool.

  Only used with the yugabyte and raft datastores.

  Example: `InterUSS`
  EOT

  default = "default_orga"
}
