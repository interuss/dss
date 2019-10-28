local dss = import "dss.libsonnet";


dss {
  metadata+: {
    data: {
      cluster: "gke_example1",
      namespace: "example1-ns",  
    }
  },
  cockroach+: {
    shouldInit: true,
    sset+: {
      dbHostnameSuffix:: "db.steeling-test.interussplatform.com",
      locality:: "steeling",
    }
  }
}