variable "dump_requests" {
  type        = bool
  default     = false
  description = <<-EOT
  Set to true to log full HTTP requests and responses of the core-service.
  It dumps sensitive information to logs; intended only for debugging and/or development.

  Example: `false`
  EOT
}
