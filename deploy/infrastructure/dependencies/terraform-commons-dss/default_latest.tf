locals {
  rid_db_schema = var.desired_rid_db_version == "latest" ? (var.datastore_type == "cockroachdb" ? "4.0.0" : "1.0.1") : var.desired_rid_db_version
  scd_db_schema = var.desired_scd_db_version == "latest" ? (var.datastore_type == "cockroachdb" ? "3.3.0" : "1.1.0") : var.desired_scd_db_version
  aux_db_schema = var.desired_aux_db_version == "latest" ? "1.1.0" : var.desired_aux_db_version
}
