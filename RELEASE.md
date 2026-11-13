# Releasing the DSS

This document is for InterUSS committers. Release notes are on the [releases page](https://github.com/interuss/dss/releases).

## Versions

Releases are git tags `interuss/dss/vX.Y.Z[-W]`, following [semantic versioning](https://semver.org).

- **Release candidates** (`vX.Y.Z-rcN`) are for early adopters and should be considered experimental because no validation is required.
- **Stable versions** (`vX.Y.Z`) are usually a promoted release candidate. Before a stable version is published, it must be validated on a DSS pool of two DSS instances hosted in two separate clouds (see [release/infrastructure](./release/infrastructure)):
  - it must pass the [prober](https://github.com/interuss/monitoring/tree/main/monitoring/prober) and the [USS qualifier](https://github.com/interuss/monitoring/tree/main/monitoring/uss_qualifier);
  - its performance shall be documented with the [benchmarker](https://github.com/interuss/monitoring/tree/main/monitoring/benchmarker).

  The [validation artifacts](#validation-artifacts) are attached to the release. A discussion page with extended validation information may also be created.

To choose the version, start from the anticipated version in [NEXT_RELEASE_NOTES.md](NEXT_RELEASE_NOTES.md).

## Publish a release candidate

1. Check that [CI](https://github.com/interuss/dss/actions/workflows/ci.yml) is green on `master`.
2. Go to [Releases → Draft a new release](https://github.com/interuss/dss/releases/new):
   - **Tag:** `interuss/dss/vX.Y.Z-rcN` on `master`. **Title:** `vX.Y.Z-rcN`.
   - Click **Generate release notes**, then paste the pending content of NEXT_RELEASE_NOTES.md at the top.
   - Select release label: `Pre-release`
   - Click **Publish release**. This pushes the tag.
3. Wait for [dss-publish.yml](.github/workflows/dss-publish.yml) to build, sign and push `docker.io/interuss/dss:vX.Y.Z-rcN`. This also updates the `latest` image tag.
4. Optionally, validate the image:
   1. Follow [release/README.md](release/README.md) with `IMAGE=docker.io/interuss/dss:vX.Y.Z-rcN`. Pin the monitoring version with `MONITORING_IMAGE=interuss/monitoring:vA.B.C` (it defaults to `latest`).
   2. Edit the release: add a **Validation** section that gives the monitoring version used and summarizes the results, then attach the artifacts.

## Promote to stable

Repeat steps 2–4 with the tag `interuss/dss/vX.Y.Z`, targeting the commit of the release candidate. Set release label as `Latest`. Validation (step 4) is mandatory.

## After a release

Open a PR on NEXT_RELEASE_NOTES.md that removes the published notes and sets the anticipated next version to the next patch (for example `v0.24.0` → `v0.24.1`).

## Releasing from a fork

1. Point your remote origin to your fork. Tags become `[owner]/dss/vX.Y.Z`.
2. In [dss-publish.yml](.github/workflows/dss-publish.yml), change the tag trigger and the job's `if:` condition to match your fork.
3. [Enable GitHub Actions](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/enabling-features-for-your-repository/managing-github-actions-settings-for-a-repository#configuring-required-approval-for-workflows-from-public-forks) in your fork.
4. Set the secrets `DOCKER_USERNAME`, `DOCKER_PASSWORD` and `DOCKER_URL`.

Alternatively, build with [build/build.sh](build/build.sh) and push the `interuss-local/dss` image to your own registry.

## Build versions

[`scripts/git/version.sh`](scripts/git/version.sh) derives the version from the latest tag. It appends `-<commit hash>` if there are commits after the tag, and `-dirty` if there are uncommitted changes. Only release versions that have neither suffix.