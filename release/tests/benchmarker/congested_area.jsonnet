local test_name = 'Congested area';
local db_type = 'crdb';
local dss_instances = ['aws', 'google'];
local users_per_step = 2;

local location = {
  horizontal: {lat: 34, lng: -118},
  vertical: {value: 300, reference: 'W84', units: 'M'},
};

local shape = {
  origin_horizontal: {lat: 0, lng: 0},
  origin_vertical: {value: 0, reference: 'W84', units: 'M'},
  origin_time: '2026-01-01T00:00:00Z',
  volumes: [
    {
      volume: {
        outline_polygon: {
          vertices: [
            {lat: -0.00001, lng: -0.00001},
            {lat: 0.00001, lng: -0.00001},
            {lat: 0.00001, lng: 0.00001},
            {lat: -0.00001, lng: 0.00001},
          ],
        },
        altitude_lower: {value: 0, reference: 'W84', units: 'M'},
        altitude_upper: {value: 20, reference: 'W84', units: 'M'},
      },
      time_start: '2026-01-01T00:00:00Z',
      time_end: '2026-01-01T00:00:05Z',
    },
  ],
};

{
  resources: {
    resource_declarations: {
      utm_auth: {
        resource_type: 'resources.communications.AuthAdapterResource',
        specification: {
          auth_spec: 'DummyOAuth(http://oauth.authority.localutm:8085/token,benchmarker)',
          scopes_authorized: [
            'utm.strategic_coordination',
          ],
        },
      },
    } + {
      ['%s_dss_pool' % dss]: {
        resource_type: 'resources.astm.f3548.v21.DSSInstancesResource',
        dependencies: {
          auth_adapter: 'utm_auth',
        },
        specification: {
          dss_instances: [
            {
              participant_id: 'dss_%s_%s' % [dss, db_type],
              base_url: 'https://dss.release-%s.ci.%s-interuss.uspace.dev' % [db_type, dss]
            },
          ],
        },
      } for dss in dss_instances
    },
  },

  actions: [
    {
      name: 'Generate intermediate artifacts',
      generate_artifacts: {
        subfolder: 'f"intermediate{action_invocation}"',
        defined_artifact_indices: [0, 1],
      },
    },
  ],

  user_types: [
    {
      name: 'FPU_%s' % dss, // Flight planner user using particular DSS instance
      flight_planner: {
        flight_generation: {
          independent_time_location_shape: {
            time: {
              fixed_spacing: '29s',
              uniform_random_spacing: '2s',
            },
            location: {
              fixed_location: location,
            },
            shape: {
              fixed_volumes: shape,
            },
          },
        },
        flight_execution: {
          end_flight_after_start: '10s',
        },
        scd_behavior: {
          dss_pool: ['%s_dss_pool' % dss],
          dss_selection_strategy: 'Random',
          subscription_strategy: {
            single_subscription: {
              subscription_id: '3bdb0b88-a522-4286-9499-160e56c953bb',
              duration: '23h',
              area: {
                lat_min: 34 - 0.00001,
                lng_min: -118 - 0.00001,
                lat_max: 34 + 0.00001,
                lng_max: -118 + 0.00001,
              },
              min_alt: {value: 0, units: 'M', reference: 'W84'},
              max_alt: {value: 3000, units: 'M', reference: 'W84'},
            },
          },
          op_intent_ref_creation_strategy: {
            ovn_coordination_group: 'cluster1',
            coordinate_requested_ovns: true,
            retries: 2,
            accept_before_flight_start: '20s',
            activate_before_flight_start: '10s',
            expect_timely_clearance: true,
          },
          op_intent_ref_cleanup_strategy: {
            after_actual_flight_end: '1s',
          },
        },
      },
    } for dss in dss_instances
  ],

  loads: [
    {
      name: 'Flight planner ramp for %s' % dss,
      user_ramp: {
        user_type: 'FPU_%s' % dss,
        initial_users: users_per_step,
        additional_users_per_step: users_per_step,
        random_seed: 1234,
        throughput_stability_criteria: {
          each_user_completed_at_least: {
            count: 1,
            operations: ['workflow.flight_planner.flight'],
          },
        },
        throughput_instability_criteria: {
          any_of: [
            {
              failures_more_than: {
                count: 30,
                operations: ['workflow.flight_planner.flight'],
              },
            },
            {
              phase_duration_at_least: '120s',
            },
            {
              average_duration_more_than: {
                duration: '60s',
                operations: ['workflow.flight_planner.flight'],
              },
            },
          ],
        },
        step_completion_criteria: {
          any_of: [
            {
              sampling_duration_at_least: '90s',
            },
            {
              completed_at_least: {
                count: 100,
                operations: ['workflow.flight_planner.flight'],
              },
            },
          ],
          sampling_duration_at_least: '10s',
          completed_at_least: {
            count: 5,
            operations: ['workflow.flight_planner.flight'],
          }
        },
        load_completion_criteria: {
          any_of: [
            {
              throughput_lower_than_peak: {
                operations: ['workflow.flight_planner.flight'],
                fraction_of_peak: 0.7,
              },
            },
          ],
        },
      },
    } for dss in dss_instances
  ],

  scenarios: std.flattenArrays([
    [
      {
        name: '%s %s' % [dss_instances[dss_index - 1], db_type],
        load: 'Flight planner ramp for %s' % dss_instances[dss_index - 1],
        [if dss_index < std.length(dss_instances) then "teardown"]: ['Generate intermediate artifacts'],
      } for dss_index in std.range(1, std.length(dss_instances))
    ]
  ]),

  artifacts: [
    {
      raw_report: {
        name: 'report_%s' % db_type,
      },
    },
    {
      matplotlib_figure: {
        name: 'scalability_curve_%s' % db_type,
        title: test_name,
        n_subfigure_rows: 1,
        n_subfigure_cols: std.length(dss_instances),
        evaluation_context: [
          {
            name: 'throughputs',
            value: '[[throughput_of_step(scenario, s, types=["workflow.flight_planner.flight"], outcomes=[True])' +
                  '  for s in completed_step_indices(scenario.steps)]' +
                  ' for scenario in report.report.scenarios]',
          },
          {
            name: 'latencies',
            value: '[[latency_of_step(scenario, s, types=["query.astm.f3548.v21.dss.createOperationalIntentReference"], outcomes=[True, False]).total_seconds() * 1000' +
                  '  for s in completed_step_indices(scenario.steps)]' +
                  ' for scenario in report.report.scenarios]',
          },
        ],
        subfigures: std.flattenArrays([
          [
            {
              title: '%s %s' % [dss_instances[dss_index - 1], db_type],
              subplots: [
                {
                  render_expr: '%d < len(report.report.scenarios)' % (dss_index - 1),
                  evaluation_context: [
                    {
                      name: 'scenario_index',
                      value: '%d' % (dss_index - 1),
                    },
                    {
                      name: 'scenario',
                      value: 'report.report.scenarios[scenario_index]',
                    },
                    {
                      name: 'scale',
                      value: '[step.load_factor for step in completed_steps(scenario.steps)]',
                    },
                    {
                      name: 'failures',
                      value: '[throughput_of_step(scenario, s, types=["workflow.flight_planner.flight"], outcomes=[False])' +
                            ' for s in completed_step_indices(scenario.steps)]',
                    },
                    {
                      name: 'usl',
                      value: 'USLFit.from_data(scale, throughputs[scenario_index])',
                    },
                  ],
                  x_axis: {
                    label: 'Flight planners',
                  },
                  y_axis: {
                    label: 'Throughput\n(Flights/s)',
                    min_value: 0,
                    max_value_expr: 'max(throughputs)',
                  },
                  y_axes: [
                    {
                      label: 'Latency\n(Create op intent ms)',
                      min_value: 0,
                      max_value_expr: 'max(latencies)',
                    },
                  ],
                  xy_plots: [
                    {
                      type: 'Line',
                      color: 'lightgray',
                      label_expr: 'f"USL: $\\\\gamma$={usl.parameters.scaling_factor:.2g} $\\\\alpha$={usl.parameters.contention_factor:.2g} $\\\\beta$={usl.parameters.coherency_factor:.2g}"',
                      x_data_expr: 'scale',
                      y_data_expr: 'list(usl.compute_throughput(scale))',
                      kwargs: {
                        zorder: -1,
                      },
                    },
                    {
                      type: 'Scatter',
                      color: 'orange',
                      label_expr: '"Latency"',
                      x_data_expr: 'scale',
                      y_data_expr: 'latencies[scenario_index]',
                      y_axis: 1,
                      kwargs: {
                        zorder: -0.9,
                      },
                    },
                    {
                      type: 'Scatter',
                      color: 'green',
                      label_expr: '"Successes"',
                      x_data_expr: 'scale',
                      y_data_expr: 'throughputs[scenario_index]',
                    },
                    {
                      type: 'Scatter',
                      color: 'red',
                      label_expr: '"Failures"',
                      x_data_expr: 'scale',
                      y_data_expr: 'failures',
                    },
                  ],
                  legend: {
                    location: 'upper left',
                    font_size: 'x-small',
                    label_spacing: 0.2,
                    border_padding: 0.2,
                  },
                },
              ],
            } for dss_index in std.range(1, std.length(dss_instances))
          ]
        ]),
      },
    },
  ],
}
