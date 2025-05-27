variable "yugabyte_zone" {
  type        = string
  description = <<-EOT
  Zone of yugabyte instances, used for partionning.

  Should be set to zone unless you're doing advanced partitionning.
  EOT

  default = "zone"
}
