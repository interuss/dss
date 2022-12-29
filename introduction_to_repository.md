# Introduction to the Repository

This document aims to provide a introduction to the repository and its structure to new contributors and developers.

## Repository structure

- The main codebase for the DSS is in `/pkg` and `/cmds`, the core organization and structure of the DSS is in these directories.
- The `/interfaces` folder contains diagrams, API specifications of the standards and other test tools that come with the DSS. This folder contains references to the ASTM standard, diagrams about remote-id test suite etc.

### Introduction to the Monitoring toolset

The [`monitoring` repository](https://github.com/interuss/monitoring) contains a set of folders containing different test suites to test different capabilities of the DSS during development and production use.

#### Prober

- The first and largest monitoring tool is the "prober" which is a full integration test of the DSS.  This tool is used during [continuous integration](.github/workflows/CI.md) for the DSS.
