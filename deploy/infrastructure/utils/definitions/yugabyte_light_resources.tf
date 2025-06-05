variable "yugabyte_light_resources" {
  type        = bool
  description = <<-EOT
  Enable light resources reservation for yugabyte instances.

  Useful for a dev cluster when you don't want to overload your kubernetes cluster.
  EOT

  default = false
}
