# InterUSS DSS implementation details

## CockroachDB
DSS instances in the same DSS cluster maintain a shared DSS Airspace
Representation (DAR) using CockroachDB for data storage and
synchronization between DSS participants.  The [CockroachDB
documentation](https://www.cockroachlabs.com/docs/stable/) is recommended
reading for performance characteristics and operational caveats. We list some of
the caveats that we've run into below:

### CockroachDB Notes
*   CockroachDB (CRDB) currently uses certificates to authenticate clients and
    node to node communication. All of Node certs, client certs, and even a CA
    cert are all generated through the cockroach cli. The CA certs must be
    custom generated since the Node certs are require to have CN=node, which no
    public CA will sign.
*   CockroachDB allows public CA certs to be concatenated, to allow for
    certificate rotation. We abuse this to allow each DSS participant to bring
    their own CA cert.
*   CockroachDB nodes join the cluster virally. That is, each node keeps state
    on all the other nodes, and if a node connects to a new node, it will learn
    about the entire cluster through a gossip protocol.
*   Each CRDB node must be uniquely addressable and routable.
    *   Because of this, we expect each node to have it's own static IP and/or
        publicly resolvable hostname. In the future, we likely want to explore
        the use of service mesh frameworks.
*   CRDB splits up data based on the locality string. Data replication strategy
    is a database variable. The CRDB nodes will traverse the key,values of the
    locality flag to determine how to divy up the replicas. This means there
    could be more than 7 participants of the DSS, with only 3 or 5 (or 7)
    replicas, and each participant would simply receive a shard of a single
    replica.
*   CRDB clients talk to any CRDB node which will proxy traffic to the correct
    node(s).
*   There is an admin UI for CRDB (default port 8080). This should be locked
    down to internal traffic only.
*   The cluster init command must *only* ever be run against one node in the
    cluster. It seeds the data directories, and if it is run against another
    node it won't be able to join the cluster, or will create some splits within
    the cluster. The command is not dangerous once a node has joined a cluster
    that has been initiated, but it is possible a node's data directory gets
    destroyed so it should be avoided.
