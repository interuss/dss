# InterUSS Platform Docker deployment

## Introduction

The contents of this folder enable the bring-up of a docker-compose system to
host an InterUSS Platform data node in a single command.

## Contents

### Dockerfile_storageapi

This Dockerfile builds an image containing the InterUSS Platform storage API. It
requires a separate Zookeeper instance to operate.

### docker-compose_storageapitest.yaml

This docker-compose configuration tests the storage API image above by
instantiating a storage API container along with a connected Zookeeper node in
stand alone mode. With this system up, the InterUSS Platform storage API is
exposed on localhost:INTERUSS_API_PORT.

### Dockerfile_reverseproxy

This Dockerfile builds an image containing an nginx reverse proxy intended to
gate requests to the storage API and provide HTTPS access to the API.

### docker-compose.yaml

This docker-compose configuration brings up an entire InterUSS Platform data
node in a single command.  By default, HTTP access to the API is available on
port 8121 and HTTPS on 8120.

### docker-compose_localssl.yaml

By layering this docker-compose configuration on top of docker-compose.yaml,
users may provide their own SSL certificates. This is the intended usage in
production.

## Usage

### Stand-alone test node

To run a stand-alone test InterUSS Platform data node that does not synchronize
with any other data nodes:

```shell
docker-compose -p datanode up
```

To verify operation, navigate a browser to http://localhost:8121/status

To make sure you have the latest versions, first run:

```shell
docker pull interussplatform/storage_api
docker pull zookeeper
docker pull interussplatform/reverse_proxy
```

### Synchronized node

To run a fully-functional non-production InterUSS Platform data node that
synchronizes with a network of other InterUSS Platform data nodes:

```shell
export ZOO_MY_ID=[your InterUSS Platform network Zookeeper ID]
export ZOO_SERVERS=[InterUSS Platform server network; ex: server.1=0.0.0.0:2888:3888 server.2=zoo2:2888:3888 server.3=zoo3:2888:3888]
export INTERUSS_PUBLIC_KEY=[paste public key here]
docker-compose -p datanode up
```

## SSL

By default, the data node docker-compose configuration will serve HTTPS
requests on port 8120 using a test self-signed certificate included in this
repository. This is insecure, and a warning will be displayed in the nginx
container. To provide a secure HTTPS connection, a different certificate must
be provided.

To generate a self-signed certificate, run this command on the host system:

```shell
openssl req -x509 -newkey rsa:4096 -nodes -out cert.pem -keyout key.pem -days 365
```

Note that self-signed certificates do not guarantee the identity of the remote
host. To be fully secure, the certificate must be signed by a trusted
certificate authority, and that certificate will only be valid on the host for
which it was signed.

To run a fully-functional production InterUSS Platform data node that
synchronizes with a network of other InterUSS Platform data nodes:

```shell
export ZOO_MY_ID=[your InterUSS Platform network Zookeeper ID]
export ZOO_SERVERS=[InterUSS Platform server network; ex: server.1=0.0.0.0:2888:3888 server.2=zoo2:2888:3888 server.3=zoo3:2888:3888]
export INTERUSS_PUBLIC_KEY=[paste public key here]
export SSL_CERT_PATH=[/path/containing/cert.pem/in/it]
export SSL_KEY_PATH=[/path/containing/key.pem/in/it]
docker docker-compose -f docker-compose.yaml -f docker-compose_localssl.yaml -p datanode up
```
