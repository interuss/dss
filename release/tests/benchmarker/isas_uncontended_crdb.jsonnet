local environment = import './library/environment.libsonnet';
local pool = import './library/crdb.libsonnet';
local isas_uncontended = import '/app/monitoring/benchmarker/configurations/interuss/netrid/isas_uncontended.libsonnet';

environment.as_ramp(environment.isa_scalability_curve(isas_uncontended(std.length(pool.dss_instances), environment.rid_resources(pool), [{name: pool.name}]), 'ISAs Uncontended %s' % pool.name), 128, 128, 15)
