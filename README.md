# USS to USS Communication and Synchronization Wrapper

This repository contains a simple and open API used for separate UAS Service
Suppliers (USS) to communicate during UAS operations, known as the InterUSS
Platform™. This flexible and distributed system is used to connect multiple USSs
operating in the same general area to share safety information while protecting
operator and consumer privacy. The system is focused on facilitating
communication amongst actively operating USSs with no details about UAS
operations stored or processed on the InterUSS Platform.

### Main Features

*   Simple, open, flexible, and scalable information interface and Optimistic
    Concurrency Control to simplify USS to USS communication and control race
    conditions.
*   Effective information sharing to enable effective, deterministic conflict
    detection for flight planning across multiple overlapping 4D volumes from
    multiple USSs.
*   Effective reconstitution of flight volumes when restarting a USS or adding a
    USS to an existing environment.

### Main Assumptions

*   Trust each USS to deconflict with known flight data.
*   Auditing is available, as all USSs can verify the authorship and addition of
    erroneous conflictions.
*   Deconfliction is simply defined as no overlapping flights in time and
    volume.

For the API specification and online test area, see
https://app.swaggerhub.com/apis/InterUSS_Platform/data_node_api.

## Sequence for USS information exchange

![Simple Sequence](assets/USS0.png)

When a USS wants to plan a flight, the “planning USS” performs the following
steps:

1.  Discover - Communicates with the InterUSS platform to discover what other
    USSs have an active operation in the specific area of flight and how to
    contact them. Flight details are not stored in the InterUSS Platform for
    both minimizing data bandwidth and maximizing privacy.
2.  Sync - If there are other USSs with active operations, the planning USS
    contacts the other USS(s) with active operations in the area and retrieves
    the operational information necessary to de-conflict the operation.
3.  Plan & Verify - Once the operation is safely planned and deconflicted, the
    USS updates the InterUSS platform with their contact information for other
    USSs to use for future discovery. This update provides final verification
    that the planning was done with the latest operational picture.
4.  Delete - Once the operation is completed, the USS removes their information
    as they do not need to be contacted in the future since the USS no longer
    has operations in that specific area of flight. All of these operations are
    secured using a standard OAuth communications flow, authorized by a
    governmental authority or, in the absence of an authority in certain parts
    of the world, industry provided OAuth solutions decided by the InterUSS
    Technical Steering Committee.

![Simple Sequence](assets/USS1.png)

For an additional examples, this [Sequence Diagram](assets/USS2.png) shows a
more complex operation with three USSs, two trying to plan at the same time.
This [Sequence Diagram](assets/USS3.png) shows a USS affecting multiple grids.
This [Sequence Diagram](assets/USS4.png) shows another USS updating one of
multiple grids during an update. And finally, this
[Sequence Diagram](assets/USS5.png) shows different examples of protection with
concurrent updates.

## Directories of Interest:
* `config/` has all of the configuration required to deploy a DSS instance to a kubernetes cluster. The README in that directory contains more information.
* `pkg/` contains all of the source code for the DSS. See the README in that directory for more information.
* `cmds/` contains entry points and docker files for the actual binaries (the `http-gateway` and `grpc-backend`)
* 

## Notes

* Currently this branch only supports the Remote ID API's. 
* The current implementation relies on CockroachDB for data storage and synchronization between DSS participants. It is recommended to read up on CockroachDB for performance characteristics and operational caveats. We list some of the caveats that we've run into below:

### CockroachDB Notes
* CockroachDB (CRDB) currently uses certificates to authenticate clients and node to node communication. All of Node certs, client certs, and even a CA cert are all generated through the cockroach cli. The CA certs must be custom generated since the Node certs are require to have CN=node, which no public CA will sign.
* CockroachDB allows public CA certs to be concatenated, to allow for certificate rotation. We abuse this to allow each DSS participant to bring their own CA cert.
* CockroachDB nodes join the cluster virally. That is, each node keeps state on all the other nodes, and if a node connects to a new node, it will learn about the entire cluster through a gossip protocol.
* Each CRDB node must be uniquely addressable and routable.
  * Because of this, we expect each node to have it's own static IP and/or publicly resolvable hostname. In the future, we likely want to explore the use of Istio, which allows multi cluster private service discovery, layers on top it's own TLS so that we could run CRDB in insecure mode, and not worry about the certificate problems (Istio can use standard CA's), and would reduce latency by removing the extra hop of the CRDB per-node Loadbalancer we are currently using.
* CRDB splits up data based on the locality string. Data replication strategy is a database variable. The CRDB nodes will traverse the key,values of the locality flag to determine how to divy up the replicas. This means there could be mroe than 7 participants of the DSS, with only 3 or 5 (or 7) replicas, and each participant would simply recieve a shard of a single replica.
* CRDB clients talk to any CRDB node which will proxy traffic to the correct node(s).
* There is an admin UI for CRDB (default port 8080). This should be locked down to internal traffic only.
* The cluster init command must *only* ever be run against one node in the cluster. It seeds the data directories, and if it is run against another node it won't be able to join the cluster, or will create some splits within the cluster. The command is not dangerous once a node has joined a cluster that has been initiated, but it is possible a node's data directory gets destroyed so it should be avoided.


## Contribution

[Contribution guidelines for this project](CONTRIBUTING.md)

Every file containing source code must include copyright and license
information. This includes any JS/CSS files that you might be serving out to
browsers. (This is to help well-intentioned people avoid accidental copying that
doesn't comply with the license.)

Apache header:

    Copyright 2018 Google Inc.

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        https://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.