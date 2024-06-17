# TOKEN-SIGNER

This service aims to provide two features:
1. Create a token within ASTM Interoperability standard
2. Sign this token using an RSA private key

## Pre-requisites

This service does not implement Authorization neither Authentication, it is expected that some other service do it before forwarding the request to this one. In our current implementation, this is done in the API Gateway.

On the "certs/" directory there should be an RSA Private Key file to be used as signing key of the tokens
A new key can be generated using openssl:
   - openssl genrsa -out private-key.pem 2048
   - openssl rsa -in private-key.pem -pubout -out public-key.pem

## Installation

To compile this project:
1. At the project root, run "go build auth"
   2. This will build the runnable "auth"
3. Run "auth"

## Use

Current implementation expects to receive an HTTP request in the format:

> GET http://<KONG URL>/route/token?scope=rid.service_provider rid.display_provider&aud=ICEA

It also expects the request to have a Header named "X-Consumer-Username", where the authenticated requester is identified

### Env

Configuration settings are set in the .env file

Some key configurations are:
- ISSUER_NAME is the name of the institution running this service
- RSA_PRIVATE_KEY_FILE is the path to the RSA private key file


