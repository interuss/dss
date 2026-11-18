local environment = import './library/environment.libsonnet';
local pool = import './library/ybdb.libsonnet';
local separated_ops = import '/app/monitoring/benchmarker/configurations/interuss/scd/separated_ops.libsonnet';

environment.as_ramp(environment.without_query_details(separated_ops(std.length(pool.dss_instances), environment.scd_resources(pool), 'Separated Operations %s' % pool.name)), 5, 5, 12)
