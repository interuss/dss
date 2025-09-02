# Troubleshooting

## Check if the CockroachDB service is exposed

Unless specified otherwise in a deployment configuration, CockroachDB
communicates on port 26257.  To check whether this port is open from Mac or
Linux, e.g.: `nc -zvw3 0.db.dss.your-region.your-domain.com 26257`.  Or, search
for a "port checker" web page/app.  Port 26257 will be open on a working
CockroachDB node.

A standard TLS diagnostic may also be run on this hostname:port combination and
all results should be valid except Trust.  Certificates are signed by
"Cockroach CA" which is not a generally-trusted CA, but this is ok.

## Accessing a CockroachDB SQL terminal

To interact with the CockroachDB database directly via SQL terminal:

```
kubectl \
  --context $CLUSTER_CONTEXT exec --namespace $NAMESPACE -it \
  cockroachdb-0 -- \
  ./cockroach sql --certs-dir=cockroach-certs/
```

## Using the CockroachDB web UI

The CockroachDB web UI is not exposed publicly, but you can forward a port to
your local machine using kubectl:

### Create a user account

Pick a username and create an account:

Access the [CockrachDB SQL terminal](#accessing-a-cockroachdb-sql-terminal) then create user with sql command

    root@:26257/rid> CREATE USER foo WITH PASSWORD 'foobar';

### Access the web UI

    kubectl -n $NAMESPACE port-forward cockroachdb-0 8080

Then go to https://localhost:8080. You'll have to ignore the HTTPS certificate
warning.
