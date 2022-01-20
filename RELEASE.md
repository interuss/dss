# Release Management

Releases of the DSS are based on git tags in the format `interuss/dss/v[0-9]+\.[0-9]+\.[0-9]+`. When either an executable or image is built from a `git` checkout of the source, the most recent tag is used as the version tag. If no such tag exists, the build system defaults to v0.0.0-[commit_hash]. If commits have been added to the tag, the commit hash is appended to the version. If the workspace is not clean, `-dirty` is appended to it. The version tag is computed by `scripts/git/version.sh`.

With that, releasing a DSS version in the canonical interuss fork requires the following steps:
- Create a release tag on master using `make tag MAJOR=X MINOR=Y PATCH=Z`. The script will push a tag (`release tag`) to the remote origin under the form of `[owner]/dss/vX.Y.Z` where `[owner]` is either the organization name or the username of the origin remote url. Official releases are `interuss/dss/v#.#.#`.
- The github workflow ([.github/workflows/dss-publish.yml](.github/workflows/dss-publish.yml)) is triggered for every new release tag. It builds and publishes the DSS image to the [official docker registry](https://hub.docker.com/repository/docker/interuss/dss).

To enable releases of DSS version in a fork, the following steps are required:
  1. Set the remote origin url of the repository of the target fork. (ie git@github.com:[owner]/dss.git)
  2. Edit in ([.github/workflows/dss-publish.yml](.github/workflows/dss-publish.yml)) the trigger to match the tags of the fork's owner as well as the job's entry condition to allow the forked repository.
  3. [Enable github actions in the forked project](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/enabling-features-for-your-repository/managing-github-actions-settings-for-a-repository#configuring-required-approval-for-workflows-from-public-forks).
  4. Configure the environment variables to setup the registry. (See instructions at the top of [.github/workflows/dss-publish.yml](.github/workflows/dss-publish.yml))

Optionally, you can manually build the DSS docker image using [build/build.sh](build/build.sh), tag accordingly the image `interuss-local/dss` and push it out to an image registry of your choice.

# CockroachDB Version

When possible we try to use the latest version major and minor of CockroachDB (v 20.2.x). CRDB 20.2.0 comes with upgrades to the underlying storage of CRDB itself using a branched version of RocksDB called Pebble.

## Backwards Compatibility

20.2.x is backwards compatible with the files written by 20.1.x and upgrades are simple. However you CANNOT downgrade back to 20.1.x as the new version updates the metadata and prevents the older version of the CRDB to be started on same files. Although 20.2.x is compatible with 20.1.x in the same cluster; it was recommended that you quickly upgrade all the remianing nodes to 20.2.x and reduce the version drift as much as possible. The possible negative effect of running a mixed cluster is the difference in features available between the nodes, you must be sure that you are not leveraging the new or deprecated features as this could negatively affect your queries.
