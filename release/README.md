This folder contains script helpers, deployment configurations, and test configurations to be run on DSS images for each release.
Scripts are not expected to be robust; in case of an error, manual intervention will be needed. They are only intended to assist someone with sufficient technical knowledge.

Credentials to access clusters are not included.

Scripts will erase existing personal configurations or workspaces named `release-aws-dss-ybdb`, `release-aws-dss-crdb`, `release-google-dss-ybdb`, or `release-google-dss-crdb`.

Certificates are always regenerated from scratch.

### Release Process

#### Ensure require CLI tools are available and set up
```shell
which gcloud aws terraform

# login with aws
aws login
aws configure --profile aws-interuss-dss

# login with gcloud
gcloud auth login

# your local path may vary for this configuration, but ensure the target projects are set correctly
cat $HOME/.config/gcloud/configurations/config_default

# could help if you encounter auth issues when other projects are setup in your local environment
unset GOOGLE_CLOUD_PROJECT
```

#### Configure environment variables
* `export IMAGE=docker.io/interuss/dss:v0.22.0`
* `export ZONE_ID=(AWS zone id)`
* `export GOOGLE_PROJECT_NAME=(google project name)`
* `export ZONE_NAME=(google zone name)`
* `export AWS_PROFILE=...` (if needed)

#### Deploy clusters
* `./release/scripts/spawn-clusters.sh`

Spawn clusters defined in infrastructure/. Config is copied into the usual 'personal/' folder and the terraform module is built.
A parallel terraform apply is then run.

After this step, Kubernetes clusters are ready.

If some issues are encountered during the deployment, it is safe to perform a `terraform apply` in the `personal/*` folders.

#### Configure clusters
* `./release/scripts/configure-clusters.sh`

Fetch the Kubernetes configuration for clusters, generate certificates (trusted between clusters), and apply certificates configuration.

After this step, services are ready to be deployed.

#### Deploy services
* `./release/scripts/deploy-services.sh`

Deploy services using Helm or Tanka. Wait for the dss /healthy endpoint to return OK.

After this step, services are ready to be tested.

#### Run tests
* `./release/scripts/run-tests.sh`

Run the prober and the qualifier against deployed services.

A local 'dummy-oauth' service is spwaned to retrive tokens.

#### Compile results
* `./release/scripts/compile-results.sh`

Zip archive containing results will be available in `release/output`.

#### Destroy clusters
* `./release/scripts/destroy-clusters.sh`

Cleanup resources by:

* Uninstalling Helm / Tanka services
* Removing Kubernetes persistent volumes
* Applying terraform destroy to release clusters

No manual cleaning operations are needed after this step.

#### Compatibility matrix
* `./release/scripts/run-compat-matrix.sh`

Test which upgrade paths are possible between released DSS versions, so that the compatibility matrix of the documentation can be updated.

The list of versions to test is the `VERSIONS` array at the top of the script, ordered from oldest to newest. The version being released must be added to it.

Any local DSS instance must be stopped first with `make stop-locally`, otherwise it interferes with the stack started by the script.

For each pair of versions A (older) to B (newer), the script starts a local stack with a single datastore shared by two `core-service` instances, one running version A and the other version B. The datastore is migrated to the latest schema of version B. Pairs where B is older than A are not tested and reported as not evaluated.

Each pair is then validated by running the prober against both instances and the USS qualifier against the pool. A pair passes only if the three runs pass.

The prober and the qualifier run from `MONITORING_IMAGE`, except for pairs involving v0.20.2 which use `V0_20_2_MONITORING_IMAGE`: those results are flagged with a footnote in the generated table.

A local `dummy-oauth` image is built if it is not already present.

The docker compose stack and the qualifier configuration are in `release/compat`.

After this step, the script prints:

* A summary matrix in the terminal
* A Markdown compatibility matrix to copy into the documentation / GitHub release notes

Container logs are available in `release/logs` and qualifier reports in `release/uss_qualifier_output/compat`.
