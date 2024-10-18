variable "crdb_region" {
  type        = string
  description = <<-EOT
    Region of your DSS instance. Regions are a high-level abstraction of a geographic region, 
    and are meant to correspond directly to the region terminology used by cloud providers. 
    Each region is broken into multiple zones. Regions are used to achieve varying survival
    goals in the face of database failure. More info at 
    https://www.cockroachlabs.com/docs/stable/multiregion-overview.

    Example: <us-east1>
  EOT

  default = ""
}
