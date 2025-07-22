locals {
  rid_db_schema = var.desired_rid_db_version == "latest" ? (var.datastore_type == "cockroachdb" ? "4.0.0" : "1.0.1") : var.desired_rid_db_version
  scd_db_schema = var.desired_scd_db_version == "latest" ? (var.datastore_type == "cockroachdb" ? "3.2.0" : "1.0.1") : var.desired_scd_db_version
}
