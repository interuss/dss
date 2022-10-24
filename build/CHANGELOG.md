# CHANGELOG

## [Unreleased]

### Added
- `cockroach.storageClass` and `prometheus.storageClass` parameters to set the k8s storage classes of these services.
  By default, they use `standard`.

### Changed
- Moved the `gateway.ipName` parameter to `gateway.gkeIngress.ipName`
