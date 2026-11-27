local environment = import './library/environment.libsonnet';
local pool = import './library/raft.libsonnet';
local typical_load = import '/app/monitoring/benchmarker/configurations/interuss/scd/typical_load.libsonnet';

environment.without_query_details(typical_load(std.length(pool.dss_instances), environment.scd_resources(pool), 'Typical Load %s' % pool.name))
