# InterUSS Platform auth server Docker deployment

## Introduction

The contents of this folder enable the bring-up of a docker-compose system to
host an InterUSS Platform auth server in a single command.

## Contents

### Dockerfile_authserver

This Dockerfile builds an image containing the InterUSS Platform auth server serving via HTTP. It is
insecure to use this auth server by itself because users authenticate with Basic authentication, so
passwords are sent in the clear without an HTTPS wrapper.

### Dockerfile_authreverseproxy

This Dockerfile builds an image containing an nginx reverse proxy intended to
gate requests to the provide HTTPS access to the auth server.

### docker-compose.yaml

This docker-compose configuration brings up an entire InterUSS Platform auth server in a single
command.  By default, HTTPS access to the server is available on port 8121.

## Running an auth server

### Resources

Before starting an auth server, a few resources must be generated.

#### Access token key pair

The auth server relies on encoding access tokens with a private key and publishing a public key with
which others may validate them.  To generate this key pair, create a new folder and, in it:

```shell
openssl genrsa -out private.pem 2048
openssl rsa -in private.pem -outform PEM -pubout -out public.pem
```

Be careful never to share private.pem.

#### Roster

The core of the auth server is to translate user credentials into access tokens.  A roster defines
the set of users, their passwords, and their respective scopes.  Create `roster.txt` in the same
folder as the key pair above and populate it with one line per user.  Each line should consist of
the user's username, their hashed password, and the scopes they are to be granted, each of those
three fields separated by commas (and be careful to eliminate whitespace).  The scopes should be
separated by spaces.  The hashed password is the SHA256 hash of
"`InterUSS Platform USERNAME PASSWORD`".  So, if user `wing.com` had password `wing`, their hashed
password would be SHA256(InterUSS Platform wing.com wing), which begins b03ed6.  To compute the
SHA256 hash on a Linux command line:

```shell
echo -n "InterUSS Platform wing.com wing" | openssl dgst -sha256
```

Or, use an online SHA256 generator like https://www.xorbin.com/tools/sha256-hash-calculator, but
this is less secure because the website may gain access to the password.

When enrolling a user, it is best to have them choose their password and only send you their hashed
password so the password itself is never stored in email servers.

An example roster.txt may look like this:

```
wing.com,b03ed640ce1aed7f1dd9558c9312918b944496937e4fe81db1a3d9968a7ee1d0,interussplatform.com_operators.read interussplatform.com_operators.write
otheruss.com,97b30e68738f0a7daebb94862da7baf4dbffa5913ac57b91ea33b33baee26573,interussplatform.com_operators.read interussplatform.com_operators.write
```

#### SSL certificate

It does not make sense to run an auth server over HTTP, so SSL certificates must be provided.
Ideally, these would come from a certificate authority, but a self-signed certificate can be
generated with these commands:

```shell
mkdir certs
mkdir private
openssl req -x509 -newkey rsa:4096 -nodes -out certs/cert.pem -keyout private/key.pem -days 365
```

Note that self-signed certificates do not guarantee the identity of the remote
host. To be fully secure, the certificate must be signed by a trusted
certificate authority, and that certificate will only be valid on the host for
which it was signed.

### Running

To run a bare-HTTP InterUSS Platform auth server from the folder containing the resources above:

```shell
export INTERUSS_AUTH_PATH=`pwd`
export SSL_CERT_PATH=`pwd`/certs
export SSL_KEY_PATH=`pwd`/private
export INTERUSS_AUTH_ISSUER=yourdomain.com
cd /path/to/this/folder
docker-compose -p authserver up
```

To verify operation, navigate a browser to https://localhost:8121/status

To make sure you have the latest versions, first run:

```shell
docker pull interussplatform/auth_server
docker pull interussplatform/auth_reverse_proxy
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
