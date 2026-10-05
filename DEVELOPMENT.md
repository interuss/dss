# Development Guidelines

## Repository structure

This repository contains multiple projects. They are kept together to simplify the dependency management. The following sections provide guidance in order to harmonize them.

### Projects Expected Structure

Each project should contain the following elements:

1. `README.md`: Introduction to the project for end users. It should describe how to configure and start it.

1. `DEVELOPER.md`: Introduction for developers. It includes development notes and references to design considerations.

1. `Makefile` including the following rules:
    - `build`
    - `lint`
    - `format`: Formats files and, if applicable, fixes linter issues. This repository uses the following formatters:
      - Go: [`gofmt`](https://pkg.go.dev/cmd/gofmt)
      - Python: [`black`](https://github.com/psf/black)
    - `test`: Runs unit tests of the project.
    - `test-e2e`: If applicable, runs tests with other components in the repository.

1. If applicable, design documentation may be captured in a `DESIGN.md` or a `design` directory.

## Changing command-line flags or commands

The binaries in [`cmds/`](cmds) (`core-service`, `db-manager migrate`, `db-manager evict`, ...) are part of the deployments.
When adding, renaming or removing a flag or command, the change usually has to be propagated to the locations below:

1. **Command documentation**: the flag reference and usage notes in [`docs/`](docs), e.g. [`docs/operations/cleanup.md`](docs/operations/cleanup.md) for `evict`.

2. **Helm chart** ([`deploy/services/helm-charts/dss`](deploy/services/helm-charts/dss)):

   - `values.yaml`
   - `values.schema.json`
   - `templates/*.yaml`

3. **Tanka** ([`deploy/services/tanka`](deploy/services/tanka)):

   - `metadata_base.libsonnet`
   - The component's `.libsonnet` (e.g. `evict.libsonnet`)
   - `examples/*/main.jsonnet`

4. **Terraform** ([`deploy/infrastructure`](deploy/infrastructure)):

   - `utils/definitions/<variable>.tf`
   - `utils/variables.py`
   - Run [`utils/generate_terraform_variables.sh`](deploy/infrastructure/utils/generate_terraform_variables.sh) to regenerate the `variables.gen.tf` and `TFVARS.gen.md` files.
   - `modules/terraform-*-dss/main.tf`
   - `dependencies/terraform-commons-dss/helm.tf`
   - `dependencies/terraform-commons-dss/tanka.tf`

5. **Local development and tests**:

   - [`build/dev/docker-compose_dss.yaml`](build/dev/docker-compose_dss.yaml) and [`build/dev/startup`](build/dev/startup), if the flag is needed to run locally.
   - Test helpers e.g. in [`test/evict/evict_helper.py`](test/evict/evict_helper.py), with a test case.

## Continuous integration

Continuous integration shall be configured to run the following targets when applicable:
- `lint`
- `build`
- `test`
- `test-e2e`
