local environment = import './library/environment.libsonnet';
local pool = import './library/ybdb.libsonnet';
local congested_area = import '/app/monitoring/benchmarker/configurations/interuss/scd/congested_area.libsonnet';

environment.as_ramp(environment.without_query_details(congested_area(std.length(pool.dss_instances), environment.scd_resources(pool), 'Congested Area %s' % pool.name)), 5, 5, 12)
