variable "yugabyte_cloud" {
  type        = string
  description = <<-EOT
  Cloud of yugabyte instances, used for partionning.

  Should be set to dss unless you're doing advanced partitionning.
  EOT

  default = "dss"
}
