local dss = import "dss.libsonnet";


dss {
  metadata: {
    cluster_name: "gke_example1",
    namespace: "example1-ns",
  },
  cockroach+: {
    shouldInit: true,
    sset+: {
      dbHostnameSuffix:: "db.steeling-test.interussplatform.com",
      locality:: "steeling",
    }
  }
}