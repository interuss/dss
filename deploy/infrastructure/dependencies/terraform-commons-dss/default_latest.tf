locals {
  rid_crdb_file_version = file("db_versions/crdb/rid.version")
  scd_crdb_file_version = file("db_versions/crdb/scd.version")
  aux_crdb_file_version = file("db_versions/crdb/aux.version")

  rid_yugabyte_file_version = file("db_versions/yugabyte/rid.version")
  scd_yugabyte_file_version = file("db_versions/yugabyte/scd.version")
  
  rid_db_schema = var.desired_rid_db_version == "latest" ? (var.datastore_type == "cockroachdb" ? local.rid_crdb_file_version : local.rid_yugabyte_file_version) : var.desired_rid_db_version
  scd_db_schema = var.desired_scd_db_version == "latest" ? (var.datastore_type == "cockroachdb" ? local.scd_crdb_file_version : local.scd_yugabyte_file_version) : var.desired_scd_db_version
  aux_db_schema = var.desired_aux_db_version == "latest" ? local.aux_crdb_file_version : var.desired_aux_db_version
}
