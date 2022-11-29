# Release Management

Releases of the DSS are based on git tags in the format `interuss/dss/v[0-9]+\.[0-9]+\.[0-9]+`, optionally suffixed with `-[0-9A-Za-z-.]+`.  This tag form follows the pattern `[owner]/[component]/[semantic version]`; see [semantic version](https://semver.org) for more information.

When either an executable or image is built from a `git` checkout of the source, the most recent tag is used as the version tag. If no such tag exists, the build system defaults to v0.0.0-[commit_hash]. If commits have been added to the tag, the commit hash is appended to the version. If the workspace is not clean, `-dirty` is appended to it. The version tag is computed by [`scripts/git/version.sh`](scripts/git/version.sh).

Releasing a DSS version requires the following steps:
- Create a release tag on master using `make tag VERSION=X.Y.Z[-W]`. The script will push a tag (`release tag`) to the remote origin under the form of `[owner]/dss/vX.Y.Z[-W]`, where
    - `[owner]` is either the organization name or the username of the origin remote url
    - `X` is the major release number
    - `Y` is the minor release number
    - `Z` is the patch number
    - (optionally) `W` is the prerelease
    - `X.Y.Z[-W]` is according to [semantic versioning](https://semver.org)
        - Note that valid examples of this form include `0.1.0`, `20.0.0`, `0.5.0-rc`, `0.5.0-1.2`
    - Official releases are `interuss/dss/v#.#.#`.
- The github workflow ([.github/workflows/dss-publish.yml](.github/workflows/dss-publish.yml)) is triggered for every new release tag. On the canonical interuss fork, it builds and publishes the DSS image to the [official docker registry](https://hub.docker.com/repository/docker/interuss/dss).

To enable releases of DSS version in a fork, the following steps are required:
  1. Set the remote origin url of the repository of the target fork. (ie git@github.com:[owner]/dss.git)
  2. Edit in ([.github/workflows/dss-publish.yml](.github/workflows/dss-publish.yml)) the trigger to match the tags of the fork's owner as well as the job's entry condition to allow the forked repository.
  3. [Enable github actions in the forked project](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/enabling-features-for-your-repository/managing-github-actions-settings-for-a-repository#configuring-required-approval-for-workflows-from-public-forks).
  4. Configure the environment variables to setup the registry. (See instructions at the top of [.github/workflows/dss-publish.yml](.github/workflows/dss-publish.yml))

Optionally, you can manually build the DSS docker image using [build/build.sh](build/build.sh), tag accordingly the image `interuss-local/dss` and push it out to an image registry of your choice.

# Historical CockroachDB version note

We try to use the latest version major and minor of CockroachDB.  For a legacy deployment with CRDB prior to v20.2.0, know that CRDB 20.2.0 comes with upgrades to the underlying storage of CRDB itself using a branched version of RocksDB called Pebble.  20.2.x is backwards compatible with the files written by 20.1.x and upgrades are simple. However you CANNOT downgrade back to 20.1.x as the new version updates the metadata and prevents the older version of the CRDB to be started on same files. Although 20.2.x is compatible with 20.1.x in the same cluster, operators should quickly upgrade all the remaining nodes to 20.2.x and reduce the version drift as much as possible. The possible negative effect of running a mixed cluster is the difference in features available between the nodes, the operator must be sure not to leverage the new or deprecated features as this could negatively affect queries.
