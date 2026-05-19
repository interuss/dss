# Monitoring

## Prerequisites

Some of these [tools](../infrastructure/index.md#prerequisites) are required to interact with monitoring services.

## Grafana / Prometheus stack

When enabled, an instance of Grafana and Prometheus are deployed along with the
core DSS services; this combination allows you to view (Grafana) metrics
(collected by Prometheus).

![Monitoring stack diagram](../assets/generated/prometheus_grafana_stack.png)

This can be enabled via:

- The `enable_monitoring` option in terraform
- The `monitoring.enabled` option in helm
- By using tanka, which always enables it

### Grafana access

To access the Grafana interface, first ensure that the appropriate
cluster context is selected (`kubectl config current-context`). Then, run the
following command:

```shell
kubectl get pod | grep grafana | awk '{print $1}' | xargs -I {} kubectl port-forward {} 3000
```

While that command is running, open a browser and navigate to
[http://localhost:3000](http://localhost:3000).

The default username is `admin` with a default password of `admin` if using tanka, or a random value in a kubernetes secret named `<release>-grafana` if using helm charts.

Example to retrieve the secret in a default 'dss' release:

```shell
kubectl get secrets/dss-grafana -o jsonpath="{.data.admin-password}" | base64 -d
```

Click the magnifying glass on the left side to select a dashboard to view.

### Prometheus access

Prometheus access is protected by a client certificate. If you need to access the web interface, you will need to import a valid client certificate in your browser.

!!! info
    For day to day usage, you don't need to access Prometheus, use Grafana instead. This is only useful for debugging.

To build a pkcs12 file from a valid client certificate (use a random password):

=== "Yugabyte"
    ```
    openssl pkcs12 -export -inkey deploy/operations/certificates-management/workspace/demo/clients/client.grafana.key -in deploy/operations/certificates-management/workspace/demo/clients/client.grafana.crt  -out /tmp/cert_key.p12
    ```

=== "CockroachDB"
    ```
    openssl pkcs12 -export -inkey build/workspace/demo/client_certs_dir/client.grafana.key -in build/workspace/demo/client_certs_dir/client.grafana.crt -out /tmp/cert_key.p12
    ```

---

Then import this file as client certificate in your browser.

* Firefox: Preferences > Privacy & Security > View Certificates > Your Certificates > Import
* Chrome: Privacy and security > Security > Manage Certificates > Import

Next time you access the interface, select the certificate you just imported.

## Prometheus Federation (Multi Cluster Monitoring)

[Prometheus Federation](https://prometheus.io/docs/prometheus/latest/federation/)
enables you to easily monitor multiple clusters of the DSS that you operate,
unifying all the metrics into a single Prometheus instance where you can build
Grafana Dashboards. Enabling Prometheus Federation is optional.


![Federation stack diagram](../assets/generated/prometheus_federation.png)

To enable it, you need to do two things:

1. Externally expose the Prometheus service of the DSS clusters.
2. Deploy a "Global Prometheus" instance to unify metrics.

### Externally Exposing Prometheus

=== "Terraform managed"

    1. Set the option `prometheus_hostname` to the value of the public hostname that will be used to access your instance.
    2. Apply changes as usual, first by running terraform, and then tanka or helm

=== "Helm managed"

    1. Set `monitoring.externalService.enabled` to `true`
    2. [Optional] Set `monitoring.externalService.ip` set to a static external IP
    3. [Optional] Set `monitoring.externalService.subnet` if you use AWS.

=== "Tanka managed"

    1. Set `expose_external` to `true`
    2. [Optional] Supply a static external IP Address to `IP`

### Deploy "Global Prometheus" instance

1. Follow guide to deploy Prometheus https://prometheus.io/docs/introduction/first_steps/
2. The scrape rules for this global instance will scrape other prometheus `/federate` endpoints and are rather simple, please look at the [example configuration](https://prometheus.io/docs/prometheus/latest/federation/#configuring-federation) as a starting point.
3. You will need to enable a client certificate, as this is used to protect the endpoint.
    * This uses the same CAs in a cluster. You can use any generated certificate with all Prometheus instances in a cluster.
    * Copy the private (`client.grafana.key`), public (`client.grafana.crt`), and CA (`ca.crt`) keys and make them available to your Prometheus instance, next to the config. Folders are:

    === "Yugabyte"
        `deploy/operations/certificates-management/workspace/demo/clients/`

    === "CockroachDB"
        `build/workspace/demo/client_certs_dir/`

<ol start="4"><li>Add encryption to your config:

```
tls_config:
    ca_file:   ca.pem
    cert_file: client.grafana.crt
    key_file:  client.grafana.key

```

</li></ol>
