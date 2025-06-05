variable "yugabyte_region" {
  type        = string
  description = <<-EOT
  Region of yugabyte instances, used for partionning.

  Should be different from others USS in a cluster.
  EOT

  default = "uss-1"
}
