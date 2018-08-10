# USS to USS Communication and Synchronization Wrapper

This repository contains a simple, open, and API used for separate UAS Service
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

![Simple Sequence](assets/USS0.png) When a USS wants to plan a flight, the
“planning USS” performs the following steps:

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
multiple grids during an update. And finally, This
[Sequence Diagram](assets/USS5.png) shows different examples of protection with
concurrent updates.

## Files of interest in this package

*   src/storage_interface.py - Zookeeper Wrapper library in python. It contains
    one class of interest: USSMetadataManager with get/set/delete operations and
    an initialization with a Zookeeper connection string.

*   src/storage_api.py - Web Service API for the Zookeeper library. It will
    start a web service and serve GET/PUT/DELETE on
    /GridCellMetaData/<z>/<x>/<y>, which wraps directly to the
    USSMetadataManager.

*   src/storage_api_test.bash - bash system test script, also shows how to start
    the server in bash.

*   src/storage_api_test.py - Python unit test for the Web Service API, also
    shows how to use the API to get, set, and delete metadata.

*   src/storage_interface_test.py - Python unit test for the Zookeeper Wrapper
    library, also shows how to use the library to get, set, and delete metadata.

*   coverage-html/index.html - interactive unit test coverage.

## Installation

*   mkdir USS2USS
*   cd USS2USS
*   (upload USS2USS-x.y.z.tar.gz)
*   tar -zxvf USS2USS-x.y.z.tar.gz
*   sudo apt install python-virtualenv python-pip
*   cd ..
*   virtualenv USSenv
*   cd USSenv
*   . bin/activate
*   pip install kazoo flask pytest python-dateutil
*   pip install requests pyjwt cryptography djangorestframework
*   ln -sf ../USS2USS/src ./src
*   export INTERUSS_PUBLIC_KEY=(The public KEY for decoding JWTs)
*   python src/storage_api.py --help
    *   For example: python src/storage_api.py -z
        "10.1.0.2:2181,10.1.0.3:2181,10.1.0.4:2181" -s 0.0.0.0 -p 8121 -v

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
