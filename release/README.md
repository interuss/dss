This folder contains script helpers, deployment configurations, and test configurations to be run on DSS images for each release.
Scripts are not expected to be robust; in case of an error, manual intervention will be needed. They are only intended to assist someone with sufficient technical knowledge.

Credentials to access clusters are not included.

Scripts will erase existing personal configurations or workspaces named `release-aws-dss-ybdb`, `release-aws-dss-crdb`, `release-google-dss-ybdb`, or `release-google-dss-crdb`.

Certificates are always regenerated from scratch.

### Release Process

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
