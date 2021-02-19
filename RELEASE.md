# Release Management

Releases of the DSS are based on git tags in the format `v[0-9]+\.[0-9]+\.[0-9]+`.
When either an executable or image is built from a `git` checkout of the source, the latest tag up to the current commit satisfying the aforementioned format is used as the current version. If no such tags exists, the build system defaults to v0.0.0.

With that, releasing a DSS version and producing release artifacts requires the following steps:
  * Create a tag `vX.Y.Z` on master and push it to the remote using `make release MAJOR=X MINOR=Y PATCH=Z`
  * Optionally, build the main docker image, tag it with `vX.Y.Z` and push it out to an image registry of your choice. The official upstream image for a given release is made available at https://hub.docker.com/repository/docker/interuss/dss.

# CockroachDB Version

When possible we try to use the latest version major and minor of CockroachDB (v 20.2.x). CRDB 20.2.0 comes with upgrades to the underlying storage of CRDB itself using a branched version of RocksDB called Pebble.

## Backwards Compatibility

20.2.x is backwards compatible with the files written by 20.1.x and upgrades are simple. However you CANNOT downgrade back to 20.1.x as the new version updates the metadata and prevents the older version of the CRDB to be started on same files. Although 20.2.x is compatible with 20.1.x in the same cluster; it was recommended that you quickly upgrade all the remianing nodes to 20.2.x and reduce the version drift as much as possible. The possible negative effect of running a mixed cluster is the difference in features available between the nodes, you must be sure that you are not leveraging the new or deprecated features as this could negatively affect your queries.