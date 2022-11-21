# USS Qualifier: Automated Testing

## Introduction

The uss_qualifier tool in this folder automates verifying compliance to requirements and interoperability of multiple USS/USSPs.

## Usage

The `uss_qualifier` tool is a synchronous executable built into the `interuss/monitoring` Docker image.  To use the `interuss/monitoring` image to run uss_qualifier, specify a working directory of `/app/monitoring/uss_qualifier` and a command of `python main.py ${OPTIONS}` -- see [`run_locally.sh`](run_locally.sh) for an example that can be run on any local system that is running the required prerequisites (documented in message printed by run_locally.sh).

The primary input accepted by uss_qualifier is the "configuration" specified with the `--config` option.  This option should be a [reference to a configuration file](configurations/README.md) that the user has constructed or been provided to test the desired system for the desired characteristics.  If testing a standard local system (DSS + dummy auth + mock USSs), the user can specify an alternate configuration reference as a single argument to `run_locally.sh` (the default configuration is `configurations.dev.local_test`).

When building a custom configuration file, consider starting from [`configurations.dev.self_contained_f3548`](configurations/dev/self_contained_f3548.yaml), as it contains all information necessary to run the test without the usage of sometimes-configuring `$ref`s and `allOf`s.  See [configurations documentation](configurations/README.md) for more information.

### Quick start

This section provides a specific set of commands to execute uss_qualifier for demonstration purposes.

1. Check out this repository: `git clone https://github.com/interuss/dss`
2. Go to repository root: `cd dss`
3. Bring up a local UTM ecosystem (DSS + dummy auth): `build/dev/run_locally.sh`
4. In a separate window, bring up a mock RID Service Provider USS: `monitoring/mock_uss/run_locally_ridsp.sh`
5. In a separate window, bring up a mock RID Display Provider USS: `monitoring/mock_uss/run_locally_riddp.sh`
6. In a separate window, bring up a mock strategic conflict detection USS: `monitoring/mock_uss/run_locally_scdsc.sh`
7. Wait until all 4 windows above stop printing new text (should take 1-2 minutes usually, or up to 15 minutes the first time)
8. In a separate window, run uss_qualifier explicitly specifying a configuration to use: `monitoring/uss_qualifier/run_locally.sh configurations.dev.local_test`

After building, uss_qualifier should take a few minutes to run and then `report.json` should appear in [monitoring/uss_qualifier](.)

At this point, uss_qualifier can be run again with a different configuration targeted at the development resources brought up in steps 3-6; for instance: `monitoring/uss_qualifier/run_locally.sh configurations.dev.self_contained_f3548`

## Architecture

* [Test suites](suites/README.md)
* [Test scenarios](scenarios/README.md) (includes test case, test step, check breakdown)
* [Test configurations](configurations/README.md)
* [Test resources](resources/README.md)
