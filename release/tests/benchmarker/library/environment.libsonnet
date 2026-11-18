{
  auth_resource_declaration: function(scopes) {
    resource_type: 'resources.communications.AuthAdapterResource',
    specification: {
      environment_variable_containing_auth_spec: 'AUTH_SPEC',
      scopes_authorized: scopes,
    },
  },

  scd_resources: function(pool) {
    resource_declarations: {
      utm_auth: $.auth_resource_declaration(['utm.strategic_coordination']),
    } + {
      ['uss%d_dss_pool' % (i + 1)]: {
        resource_type: 'resources.astm.f3548.v21.DSSInstancesResource',
        dependencies: {
          auth_adapter: 'utm_auth',
        },
        specification: {
          dss_instances: [pool.dss_instances[i]],
        },
      } for i in std.range(0, std.length(pool.dss_instances) - 1)
    },
  },

  rid_resources: function(pool) {
    resource_declarations: {
      utm_auth: $.auth_resource_declaration(['rid.service_provider', 'rid.display_provider']),
    } + {
      ['dss_pool_%d' % (i + 1)]: {
        resource_type: 'resources.astm.f3411.DSSInstancesResource',
        dependencies: {
          auth_adapter: 'utm_auth',
        },
        specification: {
          dss_instances: [
            {
              participant_id: pool.dss_instances[i].participant_id,
              base_url: pool.dss_instances[i].base_url + '/rid/v2',
              rid_version: 'F3411-22a',
            },
          ],
        },
      } for i in std.range(0, std.length(pool.dss_instances) - 1)
    },
  },

  without_query_details: function(config) config + {
    scenarios: [s + {record_query_details: false} for s in config.scenarios],
    artifacts: [a for a in config.artifacts if !std.objectHas(a, 'usslogset')],
  },
}
