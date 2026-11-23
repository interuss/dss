local environment = import './library/environment.libsonnet';
local pool = import './library/raft.libsonnet';
local isas_uncontended = import '/app/monitoring/benchmarker/configurations/interuss/netrid/isas_uncontended.libsonnet';

isas_uncontended(std.length(pool.dss_instances), environment.rid_resources(pool), [{name: pool.name}])
