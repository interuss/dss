#!/usr/bin/python3

import requests


## get token ##
params = (
    ('sub', 'uss1'),
    ('intended_audience', 'uss2'),
    ('scope', 'dss.read.identification_service_areas'),
    ('issuer', 'dummy_oauth'),
)

response = requests.get('http://localhost:8085/token', params=params)

print(response.text)


