# Release Management

Releases of the DSS are based on git tags in the format `v[0-9]+\.[0-9]+\.[0-9]+`.
When either an executable or image is built from a `git` checkout of the source, the latest tag up to the current commit satisfying the aforementioned format is used as the current version. If no such tags exists, the build system defaults to v0.0.0.

With that, releasing a DSS version and producing release artifacts requires the following steps:
  * Create a tag `vX.Y.Z` on master and push it to the remote using `make release MAJOR=X MINOR=Y PATCH=Z`
  * Optionally, build the main docker image, tag it with `vX.Y.Z` and push it out to an image registry of your choice. The official upstream image for a given release is made available at https://hub.docker.com/repository/docker/interuss/dss.