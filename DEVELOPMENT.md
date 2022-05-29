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

## Continuous integration

Continuous integration shall be configured to run the following targets when applicable:
- `lint`
- `build`
- `test`
- `test-e2e`
