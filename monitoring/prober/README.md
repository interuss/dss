# Prober (integration tests)

This directory contains integration tests that can be run against a live DSS
instance.  These integration tests are also run as part of
[docker_e2e.sh](../../test/docker_e2e.sh) which is a self-contained system
integration test suite.

## Authorization
When running this prober pytest suite, you must describe how to obtain access
tokens during the tests.  There are three different circumstances under which an
access token is needed, each with a command line flag to specify how to obtain
access tokens in that circumstance:

* Performing remote ID operations (`rid-auth` flag).  This user should have
  permission to be granted all remote ID scopes.
* Performing strategic conflict detection operations as USS1 (`scd-auth1`
  flag).  This user should have permission to be granted all SCD scopes.
* Performing strategic conflict detection operations as USS2 (`scd-auth2`
  flag).  This user should have permission to perform strategic coordination,
  and the credentials must be different than those for `scd-auth1`.

The value for each of these flags is a specification for how to retrieve access
tokens.  See [the auth spec documentation](../monitorlib/README.md#Auth_specs)
for the format of these values.  Note that if an authorization spec is omitted,
the tests that depend on that authorization will be skipped. 

Example: `--rid-auth "UsernamePassword(https://example.com/token, username=uss1,
password=uss1, client_id=uss1)"`

## Running pytest via Docker
This approach takes slightly longer to execute due to building the prober image,
but it requires no setup and generally obtains more reproducible results than
running locally.  From the root of this repo:

```shell script
docker run --rm $(docker build -q -f monitoring/prober/Dockerfile monitoring) \
    --dss-endpoint <URL> \
    [--rid-auth <SPEC>] \
    [--scd-auth1 <SPEC>] \
    [--scd-auth2 <SPEC>]
```

Or, execute the two steps separately.  First, build the prober image:

(from [`monitoring`](../) working directory)
```shell script
docker build -f prober/Dockerfile . -t local-prober
```

...then run it:

```shell script
docker run --rm local-prober \
    --dss-endpoint <URL> \
    [--rid-auth <SPEC>] \
    [--scd-auth1 <SPEC>] \
    [--scd-auth2 <SPEC>]
```

## Running pytest locally
This approach will save the time of building the prober Docker container, but
requires more setup and may not be fully-reproducible due to system differences
in Python, etc.

### First time set up
```shell script
sudo apt-get install virtualenv
virtualenv --python python3 env
. ./env/bin/activate
pip install -r requirements.txt
```

### Running the prober
```shell
. ./env/bin/activate
pytest \
    --dss-endpoint <URL> \
    [--rid-auth <SPEC>] \
    [--scd-auth1 <SPEC>] \
    [--scd-auth2 <SPEC>] \
    -vv .
```

## Writing prober tests
This project strives to make prober tests as easy as possible to create in order
to encourage creation of more tests which improve the validity and reliability
of the system.  However, there are still some guidelines to follow when making
new tests.

### Creating a new test
When creating a new test, the first decision is where to locate the test.  The
prober is divided by subsystems with each subsystem tested in its respective
folder ([`aux`](aux), [`rid`](rid), [`scd`](scd)).  Each test_*.py file in those
folders contains a group of tests, sometimes necessary to be executed in
sequence.  An individual test reproducing a problem with a specific query might
be placed in test_<RESOURCE>_special_cases.py.  Otherwise, the new test should
be located in the most applicable existing group (test_*.py file), or else in a
new test file if none of the existing files are applicable.

### Test signature
Tests are as defined by [pytest](https://docs.pytest.org/en/stable/).
Generally, this means that a new test should be a function that starts with the
prefix "test_" and makes `assert` statements.  Any test fixtures needed will be
automatically (magically?) provided by pytest according to name of the arguments
to the "test_" function.  The available fixtures are defined in
[conftest.py](conftest.py) and decorated with `@pytest.fixture`.

### Making requests
To send a request to the DSS during a prober test, use one of the *session
fixtures.  They come pre-configured so that the URL specified starts at the
session domain level, so the full URL does not need to be known or specified.
See other tests for illustrations of what URL segments are necessary.

Every request made to the DSS using a session that includes authorization must
specify a scope to be included in the access token.  The session fixtures
automatically acquire and manage access tokens, but the necessary scope must
still be specified in each request.  To instruct the session to use the same
scope for all requests within a test, simply decorate the test with
`@default_scope(<SCOPE>)` or `@default_scopes(<SCOPE1>, <SCOPE2>)` (located in
[infrastructure.py](../monitorlib/infrastructure.py)).  To specify scope for an individual
request (which also overrides the test-default scope), add a `scope=<SCOPE>`
keyword argument to the request call (e.g.,
`response = session.get('/resource', scope='read')`).

### Resources
Each test file should be robust to statefulness in a DSS instance.
Specifically:

1. Each test file should ensure that no resources are left behind in the DSS
   after all the tests in the test file complete successfully.
1. Each test file interacting with resources should have the first test in the
   file delete any pre-existing resource that would interfere with the test.
1. Static IDs should be used for all test resources.  These static IDs can be
   defined at the module (test file) level as constants.
1. Static IDs should, whenever possible, clearly stand out as test IDs.  For
   UUID-formatted IDs, the suggestion is to replace the first and last 6 digits
   of the ID with zeros.  Tests that depend on ID value for their behavior do
   not need to follow this guideline.  Tests built from a failing query should
   try changing the ID to stand out as a test ID, and keep that change as long
   as the behavior illustrated does not change.
1. Tests should be designed to pass even with pre-existing resources which may
   even be in the same area/time range/etc as the test resources.  Toward this
   end, checking the exact count of results returned by a query, or expecting a
   single result from a query, should be avoided.

The guidelines above should allow the prober to be used on production DSS
instances without any nominal adverse effects.
