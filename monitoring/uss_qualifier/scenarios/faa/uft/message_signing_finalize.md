# Finalize message signing test scenario

This test scenario instructs a mock USS to finalize a message signing report from captured data.

## Resources
The current specification used for http message signing checks can be found [here](https://datatracker.ietf.org/doc/html/draft-ietf-httpbis-message-signatures-11).

### mock_uss

The means to communicate with the mock USS that has been collecting message signing data.

## Finalize message signing test case

### Signal mock USS test step

#### Successful finalization check

If the mock USS doesn't finalize the message signing report successfully, this check will fail.

#### All message signing headers present check

The first step of message signing analysis is to check for the existence of all of the required message signing headers. These headers are `content-digest`, `utm-message-signature`, `utm-message-signature-input`, and `x-utm-jws-header`. If any of these headers are missing in incoming requests to the mock_uss, or are missing in received responses from the USS under test, then that will generate an issue.

#### Valid content digests check

The sha-512 hash on either the request/response body (depending on what is being analyzed), is taken, and this value is compared to the one received in the `content-digest` header. If these values differ, this will generate an issue showing the difference in the values. 

#### Valid signature check

The `utm-message-signature` is analyzed to ensure that it is a valid signature. To do this check, the signature base is created, and the public key served in the `x5u` field within the `x-utm-jws-header` is retreived. Using the value in the `utm-message-signature` header field, the public key, and the recreated signature base, it is determined whether or not the signature is valid.