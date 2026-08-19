# DSS upgrades

This page describes how to upgrade a DSS pool, and which versions may temporarily coexist
while the upgrade is in progress.

## Upgrade preparation

Before upgrading a pool:

1. Read the release notes of every version between the currently deployed one and the target
   one, not only those of the target version.
2. Check the compatibility table below for the source and target versions.
3. Choose the procedure: a [rolling upgrade](#rolling-upgrade-procedure) if the two versions are compatible; a [Zero-Traffic upgrade](#zero-traffic-upgrade-procedure)
   otherwise, or if incompatible settings must be changed.
4. Agree on which DSS instance in the pool applies the database migrations.


## Compatibility between releases

Releases are generally backward compatible with at least the previous release (and often earlier releases as well; see below), which allows certain combinations of versions of the DSS to
run simultaneously during an upgrade.

To limit the impact of a partially upgraded pool, we recommend keeping the window during
which multiple versions run as short as practically possible.

### Versions compatibility

The following matrix shows what is possible when a user wants to upgrade a pool on version A (rows) to
version B (columns).

The table always assumes a migration to the latest schema of the target version B prior to DSS version upgrade per "Rolling upgrade procedure" below.  Where this cannot be accomplished (e.g., DSS version X cannot function with the latest schema of DSS version X+1), the transition will be indicated as incompatible.

| A \ B | [v0.20.2](https://github.com/interuss/dss/releases/tag/interuss%2Fdss%2Fv0.20.2) | [v0.21.1](https://github.com/interuss/dss/releases/tag/interuss%2Fdss%2Fv0.21.1) | [v0.22.0](https://github.com/interuss/dss/releases/tag/interuss%2Fdss%2Fv0.22.0) | [v0.23.0-rc3](https://github.com/interuss/dss/releases/tag/interuss%2Fdss%2Fv0.23.0-rc3) |
|---|---|---|---|---|
| **[v0.20.2](https://github.com/interuss/dss/releases/tag/interuss%2Fdss%2Fv0.20.2)** | ✅ | ✅<sup>1</sup> | ✅<sup>1</sup> | ✅<sup>1</sup> |
| **[v0.21.1](https://github.com/interuss/dss/releases/tag/interuss%2Fdss%2Fv0.21.1)** | ⚪ | ✅ | ✅ | ✅ |
| **[v0.22.0](https://github.com/interuss/dss/releases/tag/interuss%2Fdss%2Fv0.22.0)** | ⚪ | ⚪ | ✅ | ✅ |
| **[v0.23.0-rc3](https://github.com/interuss/dss/releases/tag/interuss%2Fdss%2Fv0.23.0-rc3)** | ⚪ | ⚪ | ⚪ | ✅ |


✅ compatible · ⚠️ degraded, see explanation below · ❌ incompatible · ⚪ not evaluated

<sup>1</sup>Some tests in a multi-version pool may fail due to improvements in the test suite and DSS behavior.

Based on these results, InterUSS suggests that a rolling upgrade may be performed between any
of the versions listed above.

### Flags compatibility

Feature flags (particularly those detailed on [the performance page](./performances.md)) are
activated per node, with no global synchronization. Modifying these flags via a progressive
rollout restart triggers a transitional state with the following impacts:

- Performance degradation: the cluster temporarily operates with mixed locking strategies. This
  untested configuration is expected to increase resource contention and response latency.
- Requirement violations: because nodes respond inconsistently during the transition, the system
  fails to comply with requirements F3548-21-DSS0210 and F3411-22a-DSS0070.

We recommend applying the [Zero-Traffic upgrade procedure](#zero-traffic-upgrade-procedure) for the following flags:

* enable_scd_global_lock
* enable_scd_hash_lock
* enable_time_based_notification_index

All flags that are local, including feature flags for monitoring, can be enabled or disabled without any specific precautions.

## Upgrade procedures

### Rolling upgrade procedure

This procedure has a lower impact and can be carried out asynchronously.

Ensure the source and target DSS versions are compatible using the table above. Periods during which
multiple DSS versions run in a DSS pool shall be as short as practically possible and limited to
transitions.

The steps are:

1. Perform the database schema migrations to the latest schema versions of the target DSS version.
2. Upgrade the DSS `core-service` instances, one at a time.

The database migrations are applied once for the whole pool, by the member agreed upon beforehand, and
before the `core-service` image can be upgraded. Each member can then upgrade their own
instances at their own pace.

### Zero-Traffic upgrade procedure

This procedure has more impact and requires more synchronization between the members of a DSS
pool. It is, however, less prone to errors, and it is required to upgrade between incompatible
releases or to change the value of settings identified above.

The steps are:

1. Stop all regular cronjobs.
2. Stop all DSS `core-service` instances.
3. Perform database migrations to the schema versions associated with the target DSS version.
4. Start the new `core-service` instances.
5. Resume all regular cronjobs.

The service is unavailable between steps 2 and 4. Every member of the pool must have completed
step 2 before step 3 can be started.
