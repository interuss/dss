import collections
import copy
import datetime
import logging
import math
import random
import requests
import threading
import time
import uuid

import formatting

EARTH_CIRCUMFERENCE = 40.075e6  # meters
ACCURACY_VERTICAL = 0.2  # meters
OPERATION_PADDING = datetime.timedelta(seconds=30)

log = logging.getLogger('Simulation')

Telemetry = collections.namedtuple('Telemetry', 'timestamp data')
LatLng = collections.namedtuple('LatLng', 'lat lng')


class Flight(object):
  def __init__(self, origin, radius, period, altitude, aircraft_info, theta0):
    self.origin = origin
    self.radius = radius
    self.period_sec = period.total_seconds()
    self.altitude = altitude
    self.aircraft_info = aircraft_info
    self._theta0 = theta0
    self.takeoff = datetime.datetime.utcnow()
    self.landing = self.takeoff + period
    self.uuid = str(uuid.uuid4())
    self.telemetry = []

  def is_flying(self):
    t = datetime.datetime.utcnow()
    return self.takeoff <= t <= self.landing

  def get_telemetry(self, history):
    t = datetime.datetime.utcnow()

    earliest = max(self.takeoff, t - history)
    positions = [t.data for t in self.telemetry if t.timestamp >= earliest]
    if not positions:
      return None
    else:
      telemetry = copy.deepcopy(self.aircraft_info['public'])
      telemetry['uuid_operation'] = self.uuid
      telemetry['position_history'] = positions
      return telemetry

  def get_bounds(self):
    dlat = self.radius / EARTH_CIRCUMFERENCE * 360
    dlng = self.radius / (
      EARTH_CIRCUMFERENCE * math.cos(math.radians(self.origin.lat))) * 360
    return (
      LatLng(self.origin.lat - dlat, self.origin.lng - dlng),
      LatLng(self.origin.lat + dlat, self.origin.lng + dlng))

  def log_telemetry(self, r):
    t = datetime.datetime.utcnow()
    if t > self.landing:
      return

    f = (t - self.takeoff).total_seconds() / self.period_sec
    p = self._location_at_fraction(f)
    self.telemetry.append(Telemetry(timestamp=t, data={
      'timestamp': formatting.timestamp(t),
      'latitude': round(p.lat, 6),
      'longitude': round(p.lng, 6),
      'height': round(self.altitude + r.normalvariate(0, ACCURACY_VERTICAL), 2),
    }))

  def _location_at_fraction(self, f):
    theta = 2 * math.pi * f + self._theta0
    x = self.radius * math.cos(theta)
    y = self.radius * math.sin(theta)
    lat = self.origin[0] + y * 360 / EARTH_CIRCUMFERENCE
    lng = self.origin[1] + x * 360 / (EARTH_CIRCUMFERENCE *
                                      math.cos(math.radians(lat)))
    return LatLng(lat=lat, lng=lng)

  def get_flight_info(self):
    info = {}
    for d in self.aircraft_info.values():
      for k, v in d.items():
        info[k] = v
    info['uuid_operation'] = self.uuid
    info['latitude_operator'] = self.origin[0]
    info['longitude_operator'] = self.origin[1]
    p = self._location_at_fraction(0)
    info['latitude_takeoff'] = round(p.lat, 6)
    info['longitude_takeoff'] = round(p.lng, 6)
    p = self._location_at_fraction(1)
    info['latitude_destination'] = round(p.lat, 6)
    info['longitude_destination'] = round(p.lng, 6)
    if len(self.telemetry) >= 2:
      t0 = self.telemetry[-2]
      t1 = self.telemetry[-1]
      dt = (t1.timestamp - t0.timestamp).total_seconds()
      info['speed_ud'] = round((t1.data['height'] - t0.data['height']) / dt, 2)
      p0 = LatLng(t0.data['latitude'], t0.data['longitude'])
      p1 = LatLng(t1.data['latitude'], t1.data['longitude'])
      dy = (p1.lat - p0.lat) * EARTH_CIRCUMFERENCE / 360
      info['speed_ns'] = round(dy / dt, 2)
      dx = ((p1.lng - p0.lng) * EARTH_CIRCUMFERENCE *
            math.cos(math.radians(p0.lat)) / 360)
      info['speed_ew'] = round(dx / dt, 2)
    return info


class FlightSim(object):
  def __init__(self, origin, radius, period, interval, min_altitude,
               max_altitude, hanger, grid_client):
    self.origin = origin
    self.radius = radius
    self.period = period
    self.interval = interval
    self.min_altitude = min_altitude
    self.max_altitude = max_altitude
    self.hanger = hanger
    self.grid_client = grid_client
    self._random = random.Random()
    self._flights = []
    self._flight_index = 0
    self._bounds = (LatLng(90, 180), LatLng(-90, -180))
    self._flightlock = threading.Lock()

    self._flightthread = threading.Thread(target=self._flightloop)
    self._flightthread.daemon = True
    self._flightthread.start()

  def _flightloop(self):
    launches = 0
    try:
      while True:
        with self._flightlock:
          t = datetime.datetime.utcnow()
          i = 0
          while i < len(self._flights):
            flight = self._flights[i]
            if t >= flight.landing + self.interval:
              del self._flights[i]
              launches += 1
            else:
              flight.log_telemetry(self._random)
              i += 1

        if launches:
          log.info('Relaunching %d flight(s)' % launches)
          for i in range(launches):
            self.launch()
          launches = 0

        time.sleep(1)
    except Exception as e:
      log.critical('Flight loop exited because ' + str(e))

  def launch(self):
    with self._flightlock:
      flight = Flight(
        self.origin, self.radius, self.period,
        self._random.uniform(self.min_altitude, self.max_altitude),
        self.hanger[self._flight_index % len(self.hanger)],
        self._random.uniform(0, 2 * math.pi))
      self._flight_index += 1
      self._flights.append(flight)
      ll, ur = flight.get_bounds()
      b = self._bounds
      self._bounds = (
        LatLng(min(ll.lat, b[0].lat), min(ll.lng, b[0].lng)),
        LatLng(max(ur.lat, b[1].lat), max(ur.lng, b[1].lng)))
      earliest = min(f.takeoff for f in self._flights) - OPERATION_PADDING
      area = self._get_area()
    latest = datetime.datetime.utcnow() + self.period + OPERATION_PADDING
    log.info('Setting operator info in grid')
    try:
      self.grid_client.set_operations(area, earliest, latest)
    except requests.exceptions.RequestException as e:
      log.error(
        'Error setting operations to prepare for launch: ' + str(e))
    log.info('Launched ' + flight.uuid)

  def land(self, i):
    with self._flightlock:
      del self._flights[i]

      if not self._flights:
        self.grid_client.remove_operations(self._get_area())
        self._bounds = (LatLng(90, 180), LatLng(-90, -180))

  def get_flights_info(self):
    with self._flightlock:
      t = datetime.datetime.utcnow()
      info = []
      for i, flight in enumerate(self._flights):
        ac = copy.deepcopy(flight.aircraft_info)
        ac['is_flying'] = flight.is_flying()
        ac['takeoff'] = round((flight.takeoff - t).total_seconds())
        ac['landing'] = round((flight.landing - t).total_seconds())
        ac['index'] = i
        ac['operation'] = flight.uuid
        info.append(ac)
      return info

  def get_flight_info(self, uuid_operation):
    with self._flightlock:
      for flight in self._flights:
        if flight.uuid == uuid_operation:
          return flight.get_flight_info()
    return None

  def get_telemetries(self, dt):
    return [ac.get_telemetry(dt) for ac in self._flights if ac.is_flying()]

  def _get_area(self):
    ll, ur = self._bounds
    area = (LatLng(ll.lat, ll.lng), LatLng(ll.lat, ur.lng),
            LatLng(ur.lat, ur.lng), LatLng(ur.lat, ll.lng))
    return area
