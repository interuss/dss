# Authentication

Every request to the core-service must include an access token verified against a set of public keys. You can provide these keys either through local files or a JWKS endpoint (the two options are mutually exclusive).

Regardless of the approach, use `accepted_jwt_audiences` to specify a comma-separated list of `aud` claims accepted by the core-service.

## Static keys

Use `public_key_files` to supply a comma-separated list of paths to PEM files containing the public keys.

These files are loaded only once at startup. Rotating a key requires a service restart.

## JWKS endpoint

To use a JWKS endpoint, set `jwks_endpoint` to the endpoint URL and `jwks_key_ids` to a comma-separated list of key IDs to extract.

The service fetches keys on startup and refreshes them every `jwks_refresh_interval`. A successful refresh replaces the entire active key set, allowing key rotations on the authorization server without restarting the core-service.

### Handling endpoint failures

A refresh fails transiently if the endpoint is unreachable, returns an error status code ($\ge 400$), or provides an invalid JWK set. In these cases, the service logs an error and continues using the last successfully retrieved keys. It will shut down only if it fails to refresh keys for longer than `jwks_key_ttl`.

Missing key IDs listed in `jwks_key_ids` are treated as non-transient errors. This signals a mismatch between the endpoint and core-service configuration, so the service shuts down immediately.

If a transient failure occurs during startup, the core-service retries with exponential backoff until the endpoint becomes available rather than crashing immediately.

`jwks_key_ttl` balances key rotation with service availability. It keeps the DSS running when the authorization server goes down, provided the key set hasn't changed. However, retired keys remain valid during this window - if a key was revoked due to a leak, an attacker with that key can continue signing accepted tokens for up to `jwks_key_ttl` (especially if they can keep the endpoint unreachable). Setting this value close to your maximum token lifespan offers a good balance.

To disable key caching entirely, set `jwks_key_ttl` to `0`.

## Development keys

For local development, the repository includes a pre-configured key pair in [`build/test-certs`](https://github.com/interuss/dss/blob/master/build/test-certs):

* `auth2.pem`: Passed to `public_key_files`
* `auth2.key`: Used by [dummy-oauth](https://github.com/interuss/dss/blob/master/cmds/dummy-oauth) to sign test tokens

!!! danger
    These keys are public. Never use them in production or on environments handling real traffic. Check the directory's README for instructions on generating new key pairs.
