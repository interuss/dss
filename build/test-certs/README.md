## About Access Tokens

This folder contains `auth2.key` and `auth2.pem` files as private and public keys access tokens that are used for validating local/development DSS instances.


[cockroach-certs](cockroach-certs/) is the directory that contains ssl certificates for the CockroachDB cluster test environment.

[messagesigning](messagesigning/) is the directory that contains public and private keys used for message signing within the [uss_qualifier](../../monitoring/uss_qualifier) test suite.

New set of access tokens can be generated using [gen-cert.go](gen-cert.go) by running following script. 

```
    docker build -t test-token-gen build/test-certs/.
    docker run -it test-token-gen
```

Generated tokens can be used as `certificate` and `private key` for .crt and .key files respectively.
