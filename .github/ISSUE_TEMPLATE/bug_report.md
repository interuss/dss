---
name: Bug report
about: Create a report to help us improve
---

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:

1. Deploy/configure DSS with '...'
2. Send request to endpoint '...' (e.g. `PUT /dss/v1/operational_intent_references/...`) with payload '...'
3. Observe HTTP status / response / log output

**Expected behavior**
A clear and concise description of what you expected to happen (e.g. expected HTTP status code, response body, or database state).

**Logs and error output**
If applicable, add relevant `core-service` logs, HTTP request/response payloads, stack traces, or `uss_qualifier` / `benchmarker` reports to help explain and reproduce your problem.


**Environment (please complete the following information):**
 - DSS Version / Commit: [e.g. `v0.23.0` or commit SHA]
 - Datastore: [e.g. `cockroachdb`, `yugabytedb`, `raftstore`]
 - Configuration Flags: [e.g. `enable_time_based_notification_index`]
 - Deployment Method: [e.g. local `docker-compose`, Helm, Tanka, Terraform]
 - Infrastructure: [e.g. AWS, GCP, Azure, on-prem, local machine]
 - Pool Size (if applicable): [e.g. single instance, 3-node instance, pool of x instances with y nodes each]

**Additional context**
Add any other context about the problem here.