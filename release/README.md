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


#### Deploy clusters (TODO)
* `./release/scripts/spawn-clusters.sh`

#### Configure clusters (TODO)
* `./release/scripts/configure-clusters.sh`

#### Deploy services (TODO)
* `./release/scripts/deploy-services.sh`

#### Run tests (TODO)
* `./release/scripts/run-tests.sh`

#### Compile results (TODO)
* `./release/scripts/compile-results.sh`

Zip archive containing results will be available in `release/output`.

#### Destroy clusters (TODO)
* `./release/scripts/destroy-clusters.sh`
