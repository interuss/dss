# Database Migrations

## Upgrading Database Schemas

All schema-related files are located in the `db_schemas` directory. Any changes you wish to make to the database schema must be done within their respective database folders. The migration files are applied in sequential numeric steps from the current version to the desired version.

By default, deployment tools target the `latest` tag and will automatically perform upgrades to the most recent version.

### Target a Specific Version

To pin your database to a specific schema version, update the corresponding deployment variables and apply the changes as you would during a standard deployment.

!!! info
    The migration tools also fully support downgrades, should the targeted version be lower than the current one.

### Manual Migrations

When operating outside of the automated deployment tools, you can use the `db-manager migrate` CLI command to apply migrations directly. Refer to the tool's built-in help (`--help`) for specific command arguments and flags.

---

!!! warning "Helm Restrictions"
    When using Helm (either directly or via Terraform), only the `latest` version will be installed for now.

=== "Terraform"
    Controlled by the following variables:

    * `desired_rid_db_version`
    * `desired_scd_db_version`
    * `desired_aux_db_version`

=== "Tanka"
    Controlled by the following variables:

    * `desired_rid_db_version`
    * `desired_scd_db_version`
    * `desired_aux_db_version`


=== "Helm"
    Manual version targeting is not available for now.
