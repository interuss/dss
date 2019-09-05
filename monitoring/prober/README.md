# Prober (integration tests)

This directory contains integration tests that can be run against a live DSS
instance.  The tests create, modify and delete subscriptions and ISAs.

### First time set up

```shell
sudo apt-get install virtualenv
virtualenv --python python3 env
. ./env/bin/activate
pip install -r requirements.txt
```

### Running the prober
Run the following command from the project's home directory:

If authenticating with a service account:

```shell
. ./env/bin/activate
pytest \
    --oauth-token-endpoint <URL> \
    --oauth-service-account-json <FILENAME> \
    --dss-endpoint <URL> \
    -vv .
```

Or if authenticating with a username/password/client_id:

```shell
docker run --rm $(docker build -q -f monitoring/prober/Dockerfile monitoring/prober) \
    --oauth-token-endpoint <URL> \
    --oauth-username <USERNAME> \
    --oauth-password <PASSWORD> \
    --oauth-client_id <CLIENT_ID> \
    --dss-endpoint <URL> \
```
