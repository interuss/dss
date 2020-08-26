// prod.libsonnet sets some production defaults, and shows how to perform overrides.
// This file shows the mimimum information required to get a DSS instance running in Kubernetes.
local dss = import '../../../deploy/dss.libsonnet';
local metadataBase = import '../../../deploy/metadata_base.libsonnet';

{
  metadata: metadataBase {
    namespace: 'dss-main',
    cockroach+: {
      shouldInit: true,
      grpc_port: 26258  # here we can override other values as well.
    },
    // We can omit other required fields here as well.
    gateway+: {
      image: 'your_image_name',
    },
    backend+: {
      image: 'your_image_name',
      traceRequests: false,
    },
  },

  all(metadata): dss.all(metadata) {
    sset+: {
      spec+: {
        template+: {
          spec+: {
            soloContainer+:: {
              command_args_+:: {
                cache: '50%',
              },
            },
          },
        },
      },
    },
  }
}
