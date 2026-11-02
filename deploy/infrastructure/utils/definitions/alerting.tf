variable "alerting" {
  type = object({
    smtp_host     = string
    smtp_email    = string
    smtp_password = string
    dest          = string
  })
  default     = null
  description = <<-EOT
  Alerting by email with Alertmanager. Alerting is disabled when not set.

  - smtp_host: SMTP server, with its port.
  - smtp_email: Email address sending the alerts, also used to authenticate on the SMTP server.
  - smtp_password: Password to authenticate on the SMTP server.
  - dest: Email address receiving the alerts.

  Example:
  ```
  {
    smtp_host     = "smtp.example.com:587"
    smtp_email    = "dss@example.com"
    smtp_password = "password"
    dest          = "team@example.com"
  }
  ```
  EOT
}
