# Sizing

## Introduction
This section contains an estimate of the computational and other resources
likely necessary to support expected demand in a country similar to the United
States.

## Time required to fulfill queries for a single flight
1. Assume 1 ISA per flight (worst case)
    1. 2 ISA management queries per flight (create & delete)
1. Assume 90% of flights are nominal and require 3 strategic deconfliction queries (Accepted, Activated, Ended) while 10% of flights have problems and require 7 strategic deconfliction queries
    1. 3.4 strategic deconfliction queries per flight
1. Assume 0.1 seconds to fulfill a query
    1. Therefore, 0.54 seconds required (on average) to fulfill management queries to support a flight

## Time required to fulfill queries for a RID Display Provider
1. Assume 2 Display Providers viewing each flight on average, 4 subscriptions per flight per DP, and 40% chance of subscription reuse
    1. 9.6 subscription queries per flight
    1. 0.96 seconds required (on average) to fulfill viewing queries to support a flight

## Required parallelism
1. Use [348,537 remote pilots in 2024](https://www.faa.gov/uas/resources/by_the_numbers/)
1. Assume 100 flights per month per remote pilot
1. Use [989,916 recreational pilots](https://www.faa.gov/data_research/aviation/aerospace_forecasts/media/FY2020-40_faa_aerospace_forecast.pdf) as a baseline (even though this is likely number of aircraft, not number of pilots) and double it for the future
1. Use [7.1 flights per month per recreational pilot](https://www.faa.gov/data_research/aviation/aerospace_forecasts/media/FY2020-40_faa_aerospace_forecast.pdf)
1. Therefore, expect about 18.6 flights per second
1. With 1.5 seconds of query time per flight, a nominal parallelism of 28 is required to satisfy the demand
1. Assuming a peak-average ratio of 3.5, a parallelism of 98 is required

## Required resources
1. With Cockroach Labs guidance of 4 parallel operations per vCPU, the DSS pool requires 25 vCPUs.
1. Assuming 3 DSS instances and the need to continue to operate when one instance is down, each DSS instance requires 13 vCPUs.
1. Using 8-vCPU virtual machines (like n2-standard-8), this means each instance needs 2 of these virtual machines
1. Assuming that 5 days' worth of flights are occupying space on disk at any given time and that each flight record on disk is 100k, approximately 83 GB of storage is required
    1. Note that Cockroach Labs recommends 4,000 read IO/s and 4,000 write IO/s, and some cloud providers scale storage speed with storage size, so 83 GB of storage may be far less than is necessary to achieve these speed numbers

## Raft store

!!! warning "Experimental"
    The Raft store is experimental: its implementation is still in progress and
    the estimates in this section have not been measured yet.

With the Raft store, every DSS node holds the data of the pool in memory,
and its Raft log and snapshots on its own disk. Every node applies every write, so the resources below are required by each node.
The pool's capacity is bounded by its smallest node.

### Memory
1. Assume 1.6 million flights per day, each lasting at most 20 minutes
    - Therefore, 23,000 flights are active on average, and 80,500 with a peak-average ratio of 3.5
2. Assume 1 ISA, 1 RID subscription, 1 operational intent and 1 SCD subscription per flight, over an area of at most 100 km²: 2 kB + 2 kB + 4 kB + 2 kB = 10 kB per flight
    - Therefore, 805 MB, rounded up to 1 GB, for the active flights
3. Double it for the metadata and the garbage collector, and double it again for the in-memory snapshot
    - Therefore, 4 GB for the data
4. The Raft log kept in memory holds at most `--raft_snapshot_interval_entries` + `--raft_snapshot_catchup_entries` entries per Raft group (15,000 by default), at about 3 kB per entry
    - Therefore, less than 50 MB per Raft group
5. With a margin of about 3x for the Go runtime, the operating system and traffic exceeding these assumptions, each node requires 16 GB of memory

### Disk
1. Assume the following Raft log entries per flight, with the same assumptions as for the memory and up to 100 conflicting flights:
    - Create and delete the ISA: 2 kB + 1 kB
    - Create and delete the RID subscription: 2 kB + 1 kB
    - 3.4 strategic deconfliction queries: 14 kB
    - Create and delete the SCD subscription: 2 kB + 1 kB
    - Therefore, about 30 kB per flight, with some margin
2. Assuming the log is compacted every hour, with 250,000 flights in the busiest hour, the write-ahead log takes up to 10 GB
3. The snapshots take up to 1 GB each, as the serialized data of the active flights
4. With a margin for the snapshots and traffic exceeding these assumptions, each node requires 100 GB of storage
    - The write-ahead log is written to disk before each write is acknowledged, so the disk latency is on the path of every write: use a local disk, or a network disk with a low fsync latency

!!! note
    Old write-ahead log files and snapshots are not deleted automatically yet:
    until they are, the disk usage keeps growing by about 30 kB per flight, so
    about 48 GB per day at 1.6 million flights per day.

### CPU
Each node requires 4 vCPUs. Every node applies every write, and the leader of each Raft group also replicates the log to the other nodes

### Network
Each node requires a 1 Gbps network. Each write is sent to every node, and a node catching up from a snapshot receives the whole data once
