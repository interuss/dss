local environment = import './library/environment.libsonnet';
local pool = import './library/crdb.libsonnet';
local separated_ops = import '/app/monitoring/benchmarker/configurations/interuss/scd/separated_ops.libsonnet';

environment.without_query_details(separated_ops(std.length(pool.dss_instances), environment.scd_resources(pool), 'Separated Operations %s' % pool.name))
