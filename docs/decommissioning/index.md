# Decommissioning a DSS instance

Before infrastructure decommissioning, the DSS instance must gracefully leave the pool or their pool components will be considered down permanently, leading to difficulty achieving quorum.

* [Leaving a CockroachDB pool](depooling-crdb.md)
* [Leaving a YugabyteDB pool](depooling-yugabyte.md)

Final decommissioning depends on how the DSS was deployed:

* [Terraform](./after-terraform.md)
* [Minikube](./minikube.md)
