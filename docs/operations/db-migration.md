## Upgrading Database Schemas

All schemas-related files are in `db_schemas` directory.  Any changes you
wish to make to the database schema should be done in their respective database
folders.  The files are applied in sequential numeric steps from the current
version M to the desired version N.

For the first-ever run during the CRDB cluster initialization, the db-manager
will run once to bootstrap and bring the database up to date.  To upgrade
existing clusters you will need to:

### If performing this operation on the original cluster
1. Update the `desired_xyz_db_version` field in `main.jsonnet`
2. Delete the existing db-manager job in your k8s cluster
3. Redeploy the newly configured db-manager with `tk apply -t job/<xyz-schema-manager>`. It should automatically up/down grade your database schema to your desired version.

### If performing this operation on any other cluster

1. Create `workspace/$CLUSTER_CONTEXT_schema_manager` in this (build) directory.

1.  From this (build) working directory,
    `cp -r ../deploy/services/tanka/examples/schema_manager/* workspace/$CLUSTER_CONTEXT_schema_manager`.

1.  Edit `workspace/$CLUSTER_CONTEXT_schema_manager/main.jsonnet` and replace all `VAR_*`
    instances with appropriate values where applicable as explained in the above section.

1.  Run `tk apply workspace/$CLUSTER_CONTEXT_schema_manager`
