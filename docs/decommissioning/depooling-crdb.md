## Leaving a pool

In an event that requires removing CockroachDB nodes we need to properly and
safely decommission to reduce risks of outages.

It is never a good idea to take down more than half the number of nodes
available in your cluster as doing so would break quorum. If you need to take
down that many nodes please do it in smaller steps.

Note: If you are removing a specific node in a Statefulset, please know that
Kubernetes does not support removal of specific node; it automatically
re-creates the node if you delete it with `kubectl delete pod`.  You will need
to scale down the Statefulset and that removes the last node first (ex:
`cockroachdb-n` where `n` is the `size of statefulset - 1`, `n` starts at 0)

1. Check if all nodes are healthy and there are no
   under-replicated/unavailable ranges:

   `kubectl exec -it cockroachdb-0 -- cockroach node status --ranges --certs-dir=cockroach-certs/`

    1. If there are under-replicated ranges changes are it is because of a node
       failure. If all nodes are healthy then it should auto recover.

    1. If there are unhealthy nodes please investigate and fix them so that the
       ranges can return to a healthy state

1. Identify the node id we intend to decommission from the previous commands
   then decommission them. The following command assumes that `cockroachdb-0` is
   not targeted for decommission otherwise select a different instance to
   connect to:

   `kubectl exec -it cockroachdb-0 -- cockroach node decommission <node id 1> [<node id 2> ...] --certs-dir=cockroach-certs/`

1. If the command executes successfully all targeted nodes should not host any
   ranges. Repeat step one to verify

    a. In the event of a hung decommission please recommission the nodes and
    repeat the above step with smaller number of nodes to decommission:

    `kubectl exec -it cockroachdb-0 -- cockroach node recommission <node id 1> [<node id 2> ...] --certs-dir=cockroach-certs/`

1. Power down the pods or delete the Statefulset, whichever is applicable

    a. Again, Statefulsets does not support deleting specific pods, as it will
       restart it immediately you will need to scale down understanding that it
       will remove node `cockroachdb-n` first; where `n` is the
       `size of statefulset - 1`.

       To proceed: `kubectl scale statefulset cockroachdb --replicas=<X>`

    b. To remove the entire Statefulset:
    `kubectl delete statefulset cockroachdb`
