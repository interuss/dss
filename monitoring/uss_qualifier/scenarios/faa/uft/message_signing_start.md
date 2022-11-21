# Start message signing test scenario

This test scenario instructs a mock USS to begin capturing message signing data.

## Resources

### mock_uss

The means to communicate with the mock USS that will collect message signing data.

## Start message signing test case

### Check mock USS readiness test step

#### Status ok check

If the mock USS doesn't respond properly to a request for its status, this check will fail.

#### Ready check

If the mock USS doesn't indicate Ready for its scd functionality, this check will fail.

### Signal mock USS test step

#### Successful start check

If the mock USS doesn't start capturing message signing data successfully, this check will fail.
