local util = import 'util.libsonnet';

{
  backendVolumes: [
    {
      name: 'public-certs',
      secret: {
        secretName: 'dss.public.certs',
        defaultMode: 256,
      },
    }
  ],
  cockroachVolumes: [
    {
      name: 'datadir',
      persistentVolumeClaim: {
        claimName: 'datadir',
      },
    },
  ],
  cockroachMounts: [
    {
      name: 'datadir',
      mountPath: '/cockroach/cockroach-data',
    },
  ],
  backendMounts: [
    {
      name: 'public-certs',
      mountPath: '/public-certs',
    },
  ],
}

