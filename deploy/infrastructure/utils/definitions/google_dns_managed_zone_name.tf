variable "google_dns_managed_zone_name" {
  type = string
  description = "GCP DNS zone name to automatically manage DNS entries. Leave it empty to manage it manually"
  default = ""
}