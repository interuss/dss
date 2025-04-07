# Monitoring

## Prerequisites

Some of the tools from [the manual deployment documentation](../../build/README.md#prerequisites) are required to interact with monitoring services.

## Grafana / Prometheus

By default, an instance of Grafana and Prometheus are deployed along with the
core DSS services; this combination allows you to view (Grafana) CRDB metrics
(collected by Prometheus).  To view Grafana, first ensure that the appropriate
cluster context is selected (`kubectl config current-context`).  Then, run the
following command:

```shell script
kubectl get pod | grep grafana | awk '{print $1}' | xargs -I {} kubectl port-forward {} 3000
```

While that command is running, open a browser and navigate to
[http://localhost:3000](http://localhost:3000).  The default username is `admin`
with a default password of `admin`.  Click the magnifying glass on the left side
to select a dashboard to view.

## Prometheus Federation (Multi Cluster Monitoring)

The DSS uses [Prometheus](https://prometheus.io/docs/introduction/overview/) to
gather metrics from the binaries deployed with this project, by scraping
formatted metrics from an application's endpoint.
[Prometheus Federation](https://prometheus.io/docs/prometheus/latest/federation/)
enables you to easily monitor multiple clusters of the DSS that you operate,
unifying all the metrics into a single Prometheus instance where you can build
Grafana Dashboards for. Enabling Prometheus Federation is optional. To enable
you need to do 2 things:
1. Externally expose the Prometheus service of the DSS clusters.
2. Deploy a "Global Prometheus" instance to unify metrics.

### Externally Exposing Prometheus
You will need to change the values in the `prometheus` fields in your metadata tuples:
1. `expose_external` set to `true`
2. [Optional] Supply a static external IP Address to `IP`
3. [Highly Recommended] Supply whitelists of [IP Blocks in CIDR form](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing), leaving an empty list mean everyone can publicly access your metrics.
4. Then Run `tk apply ...` to deploy the changes on your DSS clusters.

### Deploy "Global Prometheus" instance
1. Follow guide to deploy Prometheus https://prometheus.io/docs/introduction/first_steps/
2. The scrape rules for this global instance will scrape other prometheus `/federate` endpoint and rather simple, please look at the [example configuration](https://prometheus.io/docs/prometheus/latest/federation/#configuring-federation).
