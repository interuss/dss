# InterUSS Platform Docker deployment

## Introduction

The contents of this folder enable the construction of a single Docker image
that may be run on a virtual machine loaded with a container OS to host an
InterUSS Platform data node.

## Contents

### Dockerfile_storageapi

This Dockerfile builds an image containing the InterUSS Platform storage API. It
requires a separate Zookeeper instance to operate.

### docker-compose_storageapitest.yaml

This docker-compose configuration tests the storage API image above by
instantiating a storage API container along with three connected Zookeeper nodes
in replicated mode. With this system up, the InterUSS Platform storage API is
exposed on localhost:INTERUSS_API_PORT.

### Dockerfile_datanode_pybase

This Dockerfile builds upon the storage API image by providing a single
concurrent Zookeeper instance. The image built from this Dockerfile encapsulates
a complete InterUSS Platform data node, but it still requires a network of other
Zookeeper instances to operate.

### docker-compose_pybasetest.yaml

This docker-compose configuration tests the data node image above by
instantiating a data node container along with two additional Zookeeper nodes.
With this system up, the InterUSS Platform storage API is exposed on
localhost:8121.

## Usage

### Stand-alone test node

To run a stand-alone test InterUSS Platform data node that does not synchronize
with any other data nodes:

```shell
docker run -e INTERUSS_PUBLIC_KEY=Unused \
  -e STORAGE_API_ARGS="-t test" \
  -p 8121:8121 \
  -d interussplatform/data_node
```

To verify operation, first make sure the container is running: `docker container
ls`

Then, navigate a browser to http://localhost:8121/status

To make sure you have the latest version, first run: `docker pull
interussplatform/data_node`

### Synchronized node

To run a fully-functional InterUSS Platform data node that synchronizes with a
network of other InterUSS Platform data nodes:

```shell
export ZOO_MY_ID=[your InterUSS Platform network Zookeeper ID]
export ZOO_SERVERS=[InterUSS Platform server network; ex: server.1=0.0.0.0:2888:3888 server.2=zoo2:2888:3888 server.3=zoo3:2888:3888]
export INTERUSS_PUBLIC_KEY=[paste public key here]
docker run -e INTERUSS_PUBLIC_KEY="${INTERUSS_PUBLIC_KEY}" \
  -e ZOO_MY_ID="${ZOO_MY_ID}" \
  -e ZOO_SERVERS="${ZOO_SERVERS}" \
  -p 8121:8121 \
  -d interussplatform/data_node
```
