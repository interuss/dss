# Provider
provider "google" {
  region  = local.region
  project = var.google_project_name
}

locals {
  region = join("-", slice(split("-", var.google_zone), 0, 2))
}