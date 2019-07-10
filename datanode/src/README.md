## Files of interest:

*   storage_interface.py - Cockroach Wrapper library in python. It contains one
    class of interest: USSMetadataManager with get/set/delete operations and an
    initialization with a connection string following [libpq connection string
    semantics]{https://www.postgresql.org/docs/current/static/libpq-connect.html#LIBPQ-CONNSTRING}.

*   uss_metadata.py - information wrapper for the actual JSON data structure.

*   storage_api.py - Web Service API for the Cockroach library. It will
    start a web service and serve GET/PUT/DELETE on
    /GridCellMetaData/<z>/<x>/<y>, which wraps directly to the
    USSMetadataManager.

*   storage_api_test.bash - bash system test script, also shows how to start
    the server in bash.

*   storage_api_test.py - Python unit test for the Web Service API, also
    shows how to use the API to get, set, and delete metadata.

*   storage_interface_test.py - Python unit test for the Cockroach Wrapper
    library, also shows how to use the library to get, set, and delete metadata.


## Installation

*   git clone https://github.com/wing-aviation/InterUSS-Platform.git
*   sudo apt install python-virtualenv python-pip
*   virtualenv USSenv
*   cd USSenv
*   . bin/activate
*   pip install flask psycopg2 pytest python-dateutil pyopenssl shapely
*   pip install requests pyjwt cryptography djangorestframework pytz
*   ln -sf ../InterUSS-Platform/datanode/src ./src
*   export INTERUSS_PUBLIC_KEY=(The public KEY for decoding JWTs)
*   python src/storage_api.py --help
    *   For example: python src/storage_api.py -c
        "host=localhost port=26257 dbname=defaultdb user=root password=" -s 0.0.0.0 -p 8121 -t
        test-instance  -v

See also the configurations described in ../docker.
