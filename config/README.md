# InterUSS Platform configuration files

## Introduction

This directory contains the steps, and some helper files to get CockroachDB setup, either creating a set of nodes from scratch, or joining nodes to an existing set of nodes.

## Caveats

* The join flag typically specifies 3-5 other nodes to join. The rest of the nodes are learned through a gossip protocol. Since we will run nodes on different networks, this set of helper files attempts to join a few local nodes, and uses a loadbalancer to find nodes on other networks.
* The advertise-addr and locality-advertise-addr flags tell each CRDB node how to make itself discoverable. For nodes on the same network we can just use a hostname or internal IP, but for nodes on different networks we need to make sure they are discoverable. There's a handful of different ways to accomplish this
  * Make every node's IP address static and external
    * More difficult to manage, and mitigates some of the benefits of kubernetes, but makes discovery simple
  * Advertise a loadbalancer IP to nodes on other networks.
    * This is what we currently do in our setup, but is not required of every participating node. The downside to this is 2 extra hops: 1) to the loadbalancer which will then round robin the request to a node X. 2) Node X to the correct node, since the LB doesn't have the context for which node to visit. 
      * CRDB supports arbitrary forwarding so this works for our use case.
  * Setup DNS forwarding or mirroring
  * Setup a VPN
  * Leverage Multicluster Istio
    * [Docs](https://istio.io/docs/concepts/multicluster-deployments/)
    * [Example](https://github.com/GoogleCloudPlatform/istio-samples/tree/master/multicluster-gke/dual-control-plane)
    * This is probably what we want to do in the long term. Again there is no requirement that every node needs to use Istio. We can have 1 set of nodes using Istio, another 2 that are on a VPN, another going through a Loadbalancer, etc.
* We are using Cockroach's [combined certs](https://www.cockroachlabs.com/docs/stable/rotate-certificates.html#why-cockroachdb-creates-a-combined-ca-certificate) so that each participant can maintain secrecy of their certs private key. This means we all need to share the certs, to create the combined set.
  * We should also practice coordinated cert rotation, say 1/month.


## Steps

1. Create a kubernetes cluster, ie: follow all of the substeps in Step 1. here: https://www.cockroachlabs.com/docs/stable/orchestrate-cockroachdb-with-kubernetes-multi-cluster.html
2. Add the new cluster namespace and context to make-cluster.py, and add any clusters you are intending to join.
3. Run `$ ./make-cluster.py`
4. Set your new load balancer ip to static. These should have been output when running make-files.py.
5. Configure firewall rules to allow ingress/egress traffic to port 26257. Also make sure that all internal traffic on this port is allowed.
