Custom yugabyte image building
==============================

To support mtls authentification, we need to build a custom image until [a fix](https://github.com/yugabyte/yugabyte-db/commit/89685fa888daca54eb3164a8c301e3bda8cf41b0) is relased into an official image.

You will need a [working Yugabyte building environment](https://docs.yugabyte.com/stable/contribute/core-database/build-from-src-almalinux/). It's recommanded to install a fresh AlmaLinux 8.

At the repository's root:

Get the latest 2025.1 release:

* `git checkout v2025.1.2.1`

Apply the patch for mTls:

* `git cherry-pick 89685fa888daca54eb3164a8c301e3bda8cf41b0`

Build yugabyte in *release* mode:

* `./yb_build.sh release`

Build release archive:

* `./yb_release`

Build docker image from the created package (version/hash may be different, check output of previous command):

* `cd docker/images/`
* `./build_docker.sh -f /root/yugabyte-db/build/yugabyte-2025.1.2.1-be316174108a9aa000ee48d5462e447e1c3a5fc9-release-clang19-centos-x86_64.tar.gz`

Tag & push (version/hash may be different, check output of previous command):

* `docker tag docker.io/yugabytedb/yugabyte:2025.1.2.1-be316174108a9aa000ee48d5462e447e1c3a5fc9 interuss/yugabyte:2025.1.2.1-interuss`
* `docker push interuss/yugabyte:2025.1.2.1-interuss`
