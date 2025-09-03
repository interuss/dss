locals {
  rid_file_version = file("../../../../build/db_schemas/rid.version")
  scd_file_version = file("../../../../build/db_schemas/scd.version")

  rid_db_schema = var.desired_rid_db_version == "latest" ? (var.datastore_type == "cockroachdb" ? local.rid_file_version : "1.0.1") : var.desired_rid_db_version
  scd_db_schema = var.desired_scd_db_version == "latest" ? (var.datastore_type == "cockroachdb" ? local.scd_file_version : "1.0.1") : var.desired_scd_db_version
  aux_db_schema = var.desired_aux_db_version == "latest" ? "1.0.0" : var.desired_aux_db_version
}
