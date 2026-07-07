# raftstore

`pkg/raftstore` contains the Raft based store implementation. It lets a service replicate its state across a cluster of DSS nodes using [etcd's `raft`](https://pkg.go.dev/go.etcd.io/raft/v3) library for consensus.

## Key assumptions

- **Single Raft group:** Each service (aux, rid, scd) is backed by exactly one Raft group (one `Consensus` instance). There is no sharding or partitioning of data across multiple groups for now.
- **Persistent log:** The Raft log and snapshots are durable: each node writes its entries to a [write-ahead log (WAL)](https://pkg.go.dev/go.etcd.io/etcd/server/v3/storage/wal) and periodically persists snapshots to disk ([consensus/storage.go](consensus/storage.go)), so a node can restart and rejoin the cluster by replaying its own WAL. The log consists of a series of Raft proposals (see the data model section below), and snapshots are serialized representations of the in-memory store.
- **In-memory data projection:** The application state machine derived from the committed log (i.e. ISAs, subscriptions, operational intents, etc.) lives in-memory. It is not persisted but rather reconstructed by replaying the log / snapshot on startup.
- **Data model:** Each service (`rid`, `scd`, `aux_`) defines its own request/response structs. During replication, they are wrapped in a [`Proposal`](consensus/proposal.go), which carries metadata (an ID, the originating node, and a timestamp for determinism) with a `RequestType` string and a request payload as a serialized `Value`. [`pkg/raftstore/store.go`](store.go) replicates the `Proposal` through Raft and, once committed, the service specific raftstore implementation interprets `RequestType`/`Value` and applies the resulting change to its own in-memory store.

## Future Work

- **Reads go through Raft:** There's no `ReadIndex`-based read path yet. Read-only requests are still proposed and committed like writes. See the `TODO` in [consensus/proposal.go](consensus/proposal.go).
- **Client-initiated config changes are not supported yet:** See the `TODO` in [consensus/consensus.go](consensus/consensus.go).
- **No background WAL cleanup:** Old, WAL segment files that are no longer needed are unlocked (`ReleaseLockTo`) but not actively deleted by a cleanup goroutine yet. See the `TODO` in [consensus/storage.go](consensus/storage.go).
