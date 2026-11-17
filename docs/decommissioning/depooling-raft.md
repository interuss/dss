# Leaving a pool

!!! warning "Experimental"
    The Raft store is experimental: its implementation is still in progress and
    the procedures in this page may change.

In an event that requires removing the nodes of a DSS instance from a pool
running the [Raft store](../deployment/pooling/raft.md), the nodes must be
properly removed from the Raft groups before the resources are destroyed, to
reduce the risks of outages.

The removal of a node is a change of the membership of the pool, agreed and
recorded by the other nodes: it is done with
[dssctl raft remove](../deployment/pooling/raft.md#removing-a-node), from
any member of the pool.  As with any membership change, it is done **one node
at a time**, and a voter cannot be removed if a Raft group would be left with
fewer than three voters.  If the pool is that small, the departing participant
cannot leave it: the pool must grow first.

## Removing the nodes of the instance

1. Check the membership of the pool from a member of another participant:
   [dssctl raft list](../deployment/pooling/raft.md#inspecting-the-membership).

1. For each node of the instance to leave, remove it from the pool:

    ```shell
    dssctl raft remove --url https://<ENDPOINT>
    ```

    The command returns once the removal is adopted by the Raft groups.  The
    removed node stops itself; it does not need to be reachable for the
    removal to succeed, so a node that is down can be removed by its URL.

1. Verify with `dssctl raft list` on a node of another participant that the
   removed nodes are gone from the members and that their IDs are in the list
   of removed IDs.

The IDs of the removed nodes are retired permanently: they can never be added
to the Raft groups again.  If the participant rejoins the pool later, its
nodes come back as new nodes, with fresh data directories and new IDs.

## Removing the CA of the participant

Once the participant has no node left in the pool, the remaining participants
should remove its CA from their pool of CAs, with the
[`dss-certs.py` script](../deployment/pooling/raft.md#prerequisites):

```shell
./dss-certs.py --name <NAME> --cluster-context $CLUSTER_CONTEXT --namespace $NAMESPACE remove-pool-ca --ca-serial=<CA_SERIAL>
./dss-certs.py --name <NAME> --cluster-context $CLUSTER_CONTEXT --namespace $NAMESPACE apply
```

`<CA_SERIAL>` is the serial of the CA, displayed by
`./dss-certs.py ... list-pool-ca`.  The CA of the participant must stay in
the bundle of the pool while it still holds a node in it: dropping it before
the last node is removed cuts the remaining nodes off from the pool, which
the pool reads as a failure rather than as a departure.  The nodes read the
CA bundle at startup, so they must be
[restarted](../deployment/pooling/raft.md#restarting-the-nodes) after the
secret is updated.

## Destroying the instance

Once the nodes are removed from the pool and the CA is dropped, the instance
can be destroyed like any other DSS instance:

* [Terraform](./terraform.md)
* [Minikube](./minikube.md)

Make sure the persistent volumes of the instance are deleted with it: they
hold the data of the nodes, which is only useful to the pool.
