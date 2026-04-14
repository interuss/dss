# Release versions

Currently, InterUSS publishes two types of release versions: release candidates and stable versions. Creation of releases is based on Commiters' appreciation.

Release candidates can be identified by their suffix `-rc*`. Those versions are published for early adopters.

The source of stable versions is usually a release candidate. 

Release notes can be found on the [releases page](https://github.com/interuss/dss/releases) of this repository. 
Keeping track of breaking changes and migration instructions is done through the [NEXT_RELEASE_NOTES.md](NEXT_RELEASE_NOTES.md) file, which is updated as features are added or modified and serves as a basis for release notes.

To be promoted to stable, a release must be first tested on a DSS Pool composed by two DSS Instances hosted in two separate clouds. The verification is made by running the [prober](https://github.com/interuss/monitoring/tree/main/monitoring/prober) and the [USS qualifier](https://github.com/interuss/monitoring/tree/main/monitoring/uss_qualifier) on both DSS Instances.
Reports are attached to releases on the [releases page](https://github.com/interuss/dss/releases).

# Release Management

Releases of the DSS are based on git tags in the format `interuss/dss/v[0-9]+\.[0-9]+\.[0-9]+`, optionally suffixed with `-[0-9A-Za-z-.]+`.  This tag form follows the pattern `[owner]/[component]/[semantic version]`; see [semantic version](https://semver.org) for more information.

When either an executable or image is built from a `git` checkout of the source, the most recent tag is used as the version tag. If no such tag exists, the build system defaults to v0.0.0-[commit_hash]. If commits have been added to the tag, the commit hash is appended to the version. If the workspace is not clean, `-dirty` is appended to it. The version tag is computed by [`scripts/git/version.sh`](scripts/git/version.sh).

## Release procedure

Releasing a DSS version requires the following steps:
- Select a release version `vX.Y.Z[-W]` appropriate for the release
  - `X` is the major release number
  - `Y` is the minor release number
  - `Z` is the patch number
  - (optionally) `W` is the prerelease
  - `X.Y.Z[-W]` is according to [semantic versioning](https://semver.org)
    - Note that valid examples of this form include `0.1.0`, `20.0.0`, `0.5.0-rc`, `0.5.0-1.2`
  - `X`, `Y`, and `Z` should be selected according to the nature of the changes included in the release
    - See [NEXT_RELEASE_NOTES.md](./NEXT_RELEASE_NOTES.md) for the minimum version increment, and look for any changes that might suggest a more substantial category of release than the intended next version currently tracked in NEXT_RELEASE_NOTES
- Create a release tag via *one* of the following methods:
  - On the InterUSS fork, click Releases -> Draft a new release
    - For **Tag**, enter `interuss/dss/vX.Y.Z` (see below for format)
    - For **Release title**, enter `vX.Y.Z` (corresponding to the tag)
    - For Release notes, click **Generate release notes**, then add:
      - any content from [NEXT_RELEASE_NOTES.md](./NEXT_RELEASE_NOTES.md) to the top of the notes
      - an additional section *Validation* linking the version of the USS qualifier used and referring to the attached assets for the associated test configuration and report
    - Assets: test configuration and report of the USS qualifier validation run
  - Create a release tag on main using `make tag VERSION=X.Y.Z[-W]`. The script will push a tag (`release tag`) to the remote origin under the form of `[owner]/dss/vX.Y.Z[-W]`, where
      - `[owner]` is either the organization name or the username of the origin remote url
      - Official releases are `interuss/dss/v#.#.#`.
      - Add the pending release notes from [NEXT_RELEASE_NOTES.md](NEXT_RELEASE_NOTES.md) to the release notes.
- The github workflow ([.github/workflows/image-publish.yml](.github/workflows/dss-publish.yml)) is triggered for every new release tag. On the canonical interuss fork, it builds and publishes the DSS image to the [official docker registry](https://hub.docker.com/repository/docker/interuss/dss).
- After completing the release, open a PR to remove the pending release notes from [NEXT_RELEASE_NOTES.md](NEXT_RELEASE_NOTES.md) and update the anticipated next release version number assuming just a bug fix (e.g., v0.18.3 -> v0.18.4)
- When a PR with a change larger than the current anticipated next release version number in [NEXT_RELEASE_NOTES.md](./NEXT_RELEASE_NOTES.md) is made, it should ideally also adjust the anticipated next release version number in NEXT_RELEASE_NOTES
  - Example 1: if the most recent release was v0.18.3, NEXT_RELEASE_NOTES indicated v0.18.4, and a PR made a change larger than a bug fix, that PR should change the number in NEXT_RELEASE_NOTES to v0.19.0
  - Example 2: if the most recent release was v1.3.0, NEXT_RELEASE_NOTES indicated v1.4.0, and a PR made a bug fix or minor change, that PR does not need to update NEXT_RELEASE_NOTES
  - Example 3: if the most recent release was v3.1.4, NEXT_RELEASE_NOTES indicated v3.1.5, and a PR made a major change, that PR should change the number in NEXT_RELEASE_NOTES to v4.0.0

## Releasing from a fork

To enable releases of DSS version in a fork, the following steps are required:
  1. Set the remote origin url of the repository of the target fork. (ie git@github.com:[owner]/dss.git)
  2. Edit in ([.github/workflows/dss-publish.yml](.github/workflows/dss-publish.yml)) the trigger to match the tags of the fork's owner as well as the job's entry condition to allow the forked repository.
  3. [Enable github actions in the forked project](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/enabling-features-for-your-repository/managing-github-actions-settings-for-a-repository#configuring-required-approval-for-workflows-from-public-forks).
  4. Configure the environment variables to setup the registry. (See instructions at the top of [.github/workflows/dss-publish.yml](.github/workflows/dss-publish.yml))

Optionally, you can manually build the DSS docker image using [build/build.sh](build/build.sh), tag accordingly the image `interuss-local/dss` and push it out to an image registry of your choice.

# Historical CockroachDB version note

We try to use the latest version major and minor of CockroachDB.  For a legacy deployment with CRDB prior to v20.2.0, know that CRDB 20.2.0 comes with upgrades to the underlying storage of CRDB itself using a branched version of RocksDB called Pebble.  20.2.x is backwards compatible with the files written by 20.1.x and upgrades are simple. However you CANNOT downgrade back to 20.1.x as the new version updates the metadata and prevents the older version of the CRDB to be started on same files. Although 20.2.x is compatible with 20.1.x in the same cluster, operators should quickly upgrade all the remaining nodes to 20.2.x and reduce the version drift as much as possible. The possible negative effect of running a mixed cluster is the difference in features available between the nodes, the operator must be sure not to leverage the new or deprecated features as this could negatively affect queries.
