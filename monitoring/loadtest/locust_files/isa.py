#!env/bin/python3

import datetime
import uuid
from locust import User, task, between
from monitoring.monitorlib import auth
from monitoring.prober import common

class ISA(User):
  wait_time = between(1, 5)
  oauth = None

  @task(1)
  def create_isa(self):
    time_start = datetime.datetime.utcnow()
    time_end = time_start + datetime.timedelta(minutes=60)

    resp = self.client.put(
      '/identification_service_areas/{}'.format(str(uuid.uuid4())),
      json={
          'extents': {
              'spatial_volume': {
                  'footprint': {
                      'vertices': common.VERTICES,
                  },
                  'altitude_lo': 20,
                  'altitude_hi': 400,
              },
              'time_start': time_start.strftime(common.DATE_FORMAT),
              'time_end': time_end.strftime(common.DATE_FORMAT),
          },
          'flights_url': 'https://example.com/dss',
        },
      header=oauth.get_header(self.host, [common.SCOPE_READ, common.SCOPE_WRITE])
    )

  def on_start(self):
    self.oauth = auth.DummyOAuth("localhost:5000/token", "fake_uss")