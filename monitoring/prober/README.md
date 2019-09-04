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

```shell
. ./env/bin/activate
pytest \
    --oauth-token-endpoint <URL> \
    --service-account-json <FILENAME> \
    --dss-endpoint <URL> \
    -vv .
```
