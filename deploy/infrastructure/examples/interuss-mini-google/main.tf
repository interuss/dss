# See ../../terraform-google-dss/variables.tf for required schema.
variable "google_cluster_context" {}
variable "dss_configuration" {}

module "terraform-google-dss" {
  source                 = "../../terraform-google-dss"
  google_cluster_context = var.google_cluster_context
  dss_configuration      = var.dss_configuration
}

