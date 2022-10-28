# Validate flight sharing test step

This step verifies that a created flight is shared properly per ASTM F3548-21 by querying the DSS for flights in the area of the flight intent, and then retrieving the details from the USS if the operational intent reference is found.  See `validate_shared_operational_intent` in [test_steps.py](test_steps.py).

## DSS response check

If the DSS does not respond properly to the query that should yield the planned flight, this check will fail.

## Operational intent shared correctly check

If a reference to the operational intent for the flight is not found in the DSS or the details cannot be retrieved from the USS, this check will fail and one of the requirements **astm.f3548.v21.USS0005** or **astm.f3548.v21.USS0105** were not met.

## Correct operational intent details check

If the operational intent details reported by the USS do not match the user's flight intent, this check will fail.
