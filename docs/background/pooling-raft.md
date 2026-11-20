# DSS Pooling using the Raft store

!!! warning "Experimental"
    The Raft store is experimental: its implementation is still in progress and
    the information in this page may change.

This document describes the specifics of using the Raft store as the data store
to accomplish [pooling](./pooling.md).

With the Raft store, there is no separate database cluster. The core-service
nodes of the DSS instances are the "database nodes" of the pool.  Each node is a
member of three Raft groups (`aux`, `rid` and `scd`), each replicating the data
of one service.  For pooling operational instructions, see
[pooling with the Raft store](../deployment/pooling/raft.md).

## How the Raft store meets pooling requirements

### "Each Raft node is addressable by every other Raft node"

Every Raft node must have its own externally-accessible `host:port` (e.g.,
`0.raft.dss-prod.example.com:2407`), given to the node with `--raft_peer_url`
(e.g., `https://0.raft.dss-prod.example.com:2407`).  The three Raft groups of a
node share a single listener on this port, which must be open to the nodes of
the other participants.  The URL of a node must be unique in the pool.

### "Each Raft node is discoverable"

The list of members of a Raft group are replicated by the group
and a node becomes a member only when an existing member adds it.

An existing member add a new node with `dssctl raft add --url <its --raft_peer_url>`.
Once the change is adopted, every member of the group knows the URL of the new node,
and the new node learns the members and the data of the group from its leader. 

### "Each Raft node accepts the certificates of every other node"

Raft nodes use mutual TLS to secure their connections, and TLS includes a
mechanism to ensure the identity of the server being contacted.  This mechanism
requires a trusted root Certificate Authority (CA) to sign a certificate
containing the public key of a particular server, so a client connecting to
that server can verify that the certificate (containing the public key) is
endorsed by the root CA as being genuine.  The pool trusts itself without any
external authority, so instead each USS acts as its own root CA.  When USS 1 is presented with certificates signed by USS 2's CA, USS 1 must know
that it can trust that certificate.  We accomplish this by exchanging all USSs'
CA public keys out-of-band in a CA bundle, and specifying that certificates
signed by any of the public keys in the CA bundle should be accepted when
considering the validity of certificates presented to establish a TLS
connection between nodes.

The certificate of a node must be valid for the host of its `--raft_peer_url`, as
verified by the peers connecting to it.

### "The Raft groups are initialized"

A Raft group is like the Ship of Theseus: it is composed of many nodes which
may all be replaced, one by one, so that a given group eventually contains none
of its original nodes.
Each Raft group is created by bootstrapping it on a single waiting node with `dssctl raft bootstrap`.
The node becomes the only member of the group, and all the other nodes join it with `dssctl raft add`.

The deployment tools can run this bootstrap automatically on the first node of
one DSS instance, with the automatic bootstrap option: it then adds the other
nodes of this instance and the listed nodes of the other participants.

## Additional requirements

These requirements must be met by every DSS instance joining an
InterUSS-compatible pool.  The deployment instructions produce a system that
complies with all these requirements, so this section may be ignored if
following those instructions.

- All Raft nodes must be run with mutual TLS, by supplying the
  `--raft_tls_cert_file`, `--raft_tls_key_file` and `--raft_tls_ca_file` flags,
  with `https` peer URLs.
    - Do not specify `--raft_insecure`.
- All the nodes of the pool must use the same `--raft_cluster_id`.
- All DSS instances in the pool should point their ntpd at the same NTP servers,
  for instance those of the [NTP Pool Project](https://www.ntppool.org/en/).
