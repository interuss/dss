# Prober (integration tests)

This directory contains integration tests that can be run against a live DSS
instance.  These integration tests are also run as part of
[docker_e2e.sh](../../test/docker_e2e.sh) which is a self-contained system
integration test suite.

## Authorization
When running this prober pytest suite, you must provide information regarding
how to obtain access tokens during the tests.  There are three different
circumstances under which an access token is needed, each with a command line
flag to specify how to obtain access tokens in that circumstance:

* Performing remote ID operations (`rid-auth` flag).  This user should have
  permission to be granted all remote ID scopes.
* Performing strategic conflict detection operations as USS1 (`scd-auth1`
  flag).  This user should have permission to be granted all SCD scopes.
* Performing strategic conflict detection operations as USS2 (`scd-auth2`
  flag).  This user should have permission to perform strategic coordination,
  and the credentials must be different than those for `scd-auth1`.

The value for each of these flags is a specification for how to retrieve access
tokens.  Each specification takes the form of `AuthAdapter(value1,value2,...)`.
All available AuthAdapters are defined in [auth.py](auth.py) and the parameters
are those accepted by the particular AuthAdapter's `__init__` constructor.  Both
ordinal and keyword (e.g., `AuthAdapter(param1=value1,param2=value2)`)
parameters are accepted.  Notes:
1. The spec must be a singular shell string unbroken by spaces and so should
   probably be wrapped in quotes on the command line.
1. All parameter types are strings.
1. String values are not delimited by any quote-like characters.
1. If an authorization spec is omitted, the tests that depend on that
   authorization will be skipped. 

### Examples

* `--rid-auth "UsernamePassword(https://example.com/token, username=uss1,
   password=uss1, client_id=uss1)"`
* `--rid-auth "ServiceAccount(https://example.com/token,
   ~/credentials/account.json)"`
* `--rid-auth "DummyOAuth(http://localhost:8085/token, sub=fake_uss)"`

## Running pytest via Docker
This approach takes slightly longer to execute due to building the prober image,
but it requires no setup and generally obtains more reproducible results than
running locally.  From the root of this repo:

```shell script
docker run --rm $(docker build -q -f monitoring/prober/Dockerfile monitoring/prober) \
    --dss-endpoint <URL> \
    [--rid-auth <SPEC>] \
    [--scd-auth1 <SPEC>] \
    [--scd-auth2 <SPEC>]
```

Or, execute the two steps separately.  First, build the prober image:

```shell script
docker build -f monitoring/prober/Dockerfile monitoring/prober -t local-prober
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
