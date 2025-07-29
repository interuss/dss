# Prerequisites

Download & install the following tools to your Linux workstation:

- [Install kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) to
  interact with kubernetes
    - Confirm successful installation with `kubectl version --client` (should
      succeed from any working directory).

- [Install Docker](https://docs.docker.com/get-docker/).
    - Confirm successful installation with `docker --version`

- If developing the DSS codebase,
  [install Golang](https://golang.org/doc/install)
    - Confirm successful installation with `go version`

- Install [minikube](https://minikube.sigs.k8s.io/docs/start/) (First step only).



## Configuration
=== "Helm"
    TBD

=== "Tanka"
    - [Install tanka](https://tanka.dev/install)
        - On Linux, after downloading the binary per instructions, run
      `sudo chmod +x /usr/local/bin/tk`
        - Confirm successful installation with `tk --version`
    - Optionally install [Jsonnet](https://github.com/google/jsonnet) if editing the jsonnet templates.

## Database
=== "CockroachDB"
    TBD

=== "Yugabyte"
    [Install CockroachDB](https://www.cockroachlabs.com/get-cockroachdb/) to generate CockroachDB certificates.

    - These instructions assume CockroachDB Core.
    - You may need to run `sudo chmod +x /usr/local/bin/cockroach` after completing the installation instructions.
    - Confirm successful installation with `cockroach version`