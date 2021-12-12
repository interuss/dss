The tests in this group validate requirements unique to the flight authorisation
service described in Article 10 of
[the U-space regulation](https://eur-lex.europa.eu/legal-content/EN/TXT/HTML/?uri=CELEX:32021R0664&from=EN#d1e1041-161-1).

## Tests

### Data validation

![Data validation sequence diagram](uas_flight_authorization.png)

This test attempts to create flights with invalid values provided for various
fields needed for a U-space flight authorisation, followed by successful flight
creation when all fields are valid.
