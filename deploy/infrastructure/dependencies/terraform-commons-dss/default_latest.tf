locals {
  rid_db_schema = var.desired_rid_db_version == "latest" ?  "4.0.0" : var.desired_rid_db_version
  scd_db_schema = var.desired_scd_db_version == "latest" ?  "3.1.0" : var.desired_scd_db_version
  latest = var.image == "latest" ? "docker.io/interuss/dss:v0.6.0" : var.image
}