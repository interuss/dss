## Leaving a pool

In an event that requires removing Yugabyte nodes we need to properly and
safely decommission to reduce risks of outages.

It is never a good idea to take down more than half the number of nodes
available in your cluster as doing so would break quorum. If you need to take
down that many nodes please do it in smaller steps.

Ensure placement info is how you want it after removal. Ensure you're not
requesting impossible placement by removing nodes, otherwise it won't be
possible to request node deletion. See the section below for placement
requirements.

Note: If you are removing a specific node in a Statefulset, please know that
Kubernetes does not support removal of specific node; it automatically
re-creates the node if you delete it with `kubectl delete pod`.  You will need
to scale down the Statefulset and that removes the last node first (ex:
`yb-tserver-n` where `n` is the `size of statefulset - 1`, `n` starts at 0)

1. Check if all nodes are healthy in the web ui.

1. Connect to a Yugabyte master and copy certs, like introduced in the previous
   section.

1. For each TServer node to be removed:

    1. Blacklist one node in your cluster.

    ``yb-admin -certs_dir_name /opt/certs/yugabyte/ -client_node_name=`hostname -f` -master_addresses yb-master-0.yb-masters.default.svc.cluster.local:7100,yb-master-1.yb-masters.default.svc.cluster.local:7100,yb-master-2.yb-masters.default.svc.cluster.local:7100 change_blacklist ADD [TSERVER_PUBLIC_HOSTNAME]``

    1. Wait for the node to be drained (no user tablet-peer or system-table-peer
       in the gui). If node is not draining, you may have placement constraints
       that prevent the removal of the node.

    1. Stop one node in your cluster.

    1. Wait until the node is marked as down and cluster will go into a
       non-healthy state then wait for recovery. When everything is green again
       proceed. Depending on settings, it may take time (15m) before the node is
       marked as dead.

    1. Remove the node:

    ``yb-admin -certs_dir_name /opt/certs/yugabyte/ -client_node_name=`hostname -f` -master_addresses yb-master-0.yb-masters.default.svc.cluster.local:7100,yb-master-1.yb-masters.default.svc.cluster.local:7100,yb-master-2.yb-masters.default.svc.cluster.local:7100 remove_tablet_server [TSERVER_ID]``

    If the command is giving you an error, data of the node may not have been
    drain correctly dues to placement constraints.

    1. Remove the node from the black list:

    ``yb-admin -certs_dir_name /opt/certs/yugabyte/ -client_node_name=`hostname -f` -master_addresses yb-master-0.yb-masters.default.svc.cluster.local:7100,yb-master-1.yb-masters.default.svc.cluster.local:7100,yb-master-2.yb-masters.default.svc.cluster.local:7100 change_blacklist REMOVE [TSERVER_PUBLIC_HOSTNAME]``

    1. Fully remove the node in your cluster.

        E.g you may delete persistent volumes.


1. For each Master node to be removed:

    1. Remove the master from the master list

    ``yb-admin -certs_dir_name /opt/certs/yugabyte/ -client_node_name=`hostname -f` -master_addresses yb-master-0.yb-masters.default.svc.cluster.local:7100,yb-master-1.yb-masters.default.svc.cluster.local:7100,yb-master-2.yb-masters.default.svc.cluster.local:7100 change_master_config REMOVE_SERVER [PUBLIC HOSTNAME] 7100``

    If the master node to be removed is the current leader, you may make it step
    down with the following command:

    ``yb-admin -certs_dir_name /opt/certs/yugabyte/ -client_node_name=`hostname -f` -master_addresses yb-master-0.yb-masters.default.svc.cluster.local:7100,yb-master-1.yb-masters.default.svc.cluster.local:7100,yb-master-2.yb-masters.default.svc.cluster.local:7100 master_leader_stepdown``

Finally, each pool participant should remove master addresses from the
`yugabyte_external_nodes` list their Yugabyte nodes will attempt to contact upon
restart and remove the CA of the participant.

!!! note
    Quick reminder for CA management:

    Remove the old CA, use `./dss-certs.sh remove-pool-ca <Certificate id>`
    Finally, apply certificates on the kubernetes cluster with
    `./dss-certs.sh apply`
