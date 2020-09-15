import datetime
from typing import Dict, List, Optional

import s2sphere
import yaml
from yaml.representer import Representer

from monitoring.monitorlib import fetch, infrastructure, rid


class FetchedISAs(fetch.Interaction):
  """Wrapper to interpret a DSS ISA query as a set of ISAs."""

  @property
  def success(self) -> bool:
    return self.error is None

  @property
  def error(self) -> Optional[str]:
    # Overall errors
    if self.status_code != 200:
      return 'Failed to search ISAs in DSS ({})'.format(self.status_code)
    if self.json_result is None:
      return 'DSS response to search ISAs was not valid JSON'

    # ISA format errors
    isa_list = self.json_result.get('service_areas', [])
    for isa in isa_list:
      if 'id' not in isa:
        return 'DSS response to search ISAs included ISA without id'
      if 'owner' not in isa:
        return 'DSS response to search ISAs included ISA without owner'

    return None

  @property
  def isas(self) -> Dict[str, rid.ISA]:
    if not self.json_result:
      return {}
    isa_list = self.json_result.get('service_areas', [])
    return {isa.get('id', ''): rid.ISA(isa) for isa in isa_list}

  @property
  def flight_urls(self) -> List[str]:
    urls = set()
    for _, isa in self.isas.items():
      if isa.flights_url is not None:
        urls.add(isa.flights_url)
    return list(urls)

  def has_different_content_than(self, other):
    if not isinstance(other, FetchedISAs):
      return True
    if self.error != other.error:
      return True
    if self.success:
      my_isas = self.isas
      other_isas = other.isas
      for id in other_isas:
        if id not in my_isas:
          return True
      for id, isa in my_isas.items():
        if id not in other_isas or isa != other_isas[id]:
          return True
    return False
yaml.add_representer(FetchedISAs, Representer.represent_dict)


def isas(utm_client: infrastructure.DSSTestSession,
         box: s2sphere.LatLngRect,
         start_time: datetime.datetime,
         end_time: datetime.datetime) -> FetchedISAs:
  area = rid.geo_polygon_string(rid.vertices_from_latlng_rect(box))
  url = '/v1/dss/identification_service_areas?area={}&earliest_time={}&latest_time={}'.format(
    area, start_time.strftime(rid.DATE_FORMAT), end_time.strftime(rid.DATE_FORMAT))
  t0 = datetime.datetime.utcnow()
  resp = utm_client.get(url, scope=rid.SCOPE_READ)

  return FetchedISAs(fetch.describe_interaction(resp, t0))


class FetchedUSSFlights(fetch.Interaction):
  @property
  def success(self) -> bool:
    return not self.errors

  @property
  def errors(self) -> List[str]:
    if self.status_code != 200:
      return ['Failed to get flights ({})'.format(self.status_code)]
    if self.json_result is None:
      return ['Flights response did not include valid JSON']
    return []

  @property
  def flights(self) -> List[rid.Flight]:
    return [rid.Flight(f) for f in self.json_result.get('flights', [])]
yaml.add_representer(FetchedUSSFlights, Representer.represent_dict)


def flights(utm_client: infrastructure.DSSTestSession,
            flights_url: str,
            area: s2sphere.LatLngRect,
            include_recent_positions: bool) -> FetchedUSSFlights:
  t0 = datetime.datetime.utcnow()
  resp = utm_client.get(flights_url, params={
    'view': '{},{},{},{}'.format(
      area.lat_lo().degrees,
      area.lng_lo().degrees,
      area.lat_hi().degrees,
      area.lng_hi().degrees,
    ),
    'include_recent_positions': 'true' if include_recent_positions else 'false',
  }, scope=rid.SCOPE_READ)
  return FetchedUSSFlights(fetch.describe_interaction(resp, t0))


class FetchedUSSFlightDetails(fetch.Interaction):
  @property
  def success(self) -> bool:
    return not self.errors

  @property
  def errors(self) -> List[str]:
    if self.status_code != 200:
      return ['Failed to get flight details ({})'.format(self.status_code)]
    if self.json_result is None:
      return ['Flight details response did not include valid JSON']
    return []
yaml.add_representer(FetchedUSSFlightDetails, Representer.represent_dict)


def flight_details(utm_client: infrastructure.DSSTestSession, flights_url: str, id: str) -> FetchedUSSFlightDetails:
  t0 = datetime.datetime.utcnow()
  resp = utm_client.get(flights_url + '/{}/details'.format(id), scope=rid.SCOPE_READ)
  result = FetchedUSSFlightDetails(fetch.describe_interaction(resp, t0))
  result['requested_id'] = id
  return result


class FetchedFlights(fetch.Interaction):
  @property
  def errors(self) -> List[str]:
    if not self.isas.success:
      return ['Failed to obtain ISAs: ' + self.isas.error]
    return []

  @property
  def isas(self) -> FetchedISAs:
    return self['dss_interaction']
yaml.add_representer(FetchedFlights, Representer.represent_dict)


def all_flights(utm_client: infrastructure.DSSTestSession,
                area: s2sphere.LatLngRect,
                include_recent_positions: bool,
                get_details: bool) -> Dict:
  isa_query = isas(utm_client, area, datetime.datetime.utcnow(), datetime.datetime.utcnow())

  uss_flight_queries: Dict[str, FetchedUSSFlights] = {}
  uss_flight_details_queries: Dict[str, FetchedUSSFlightDetails] = {}
  for flights_url in isa_query.flight_urls:
    flights_for_url = flights(utm_client, flights_url, area, include_recent_positions)
    uss_flight_queries[flights_url] = flights_for_url

    if get_details and flights_for_url.success:
      for flight in flights_for_url.flights:
        if flight.valid:
          details = flight_details(utm_client, flights_url, flight.id)
          uss_flight_details_queries[flight.id] = details

  return FetchedFlights({
    'dss_isa_query': isa_query,
    'uss_flight_queries': uss_flight_queries,
    'uss_flight_details_queries': uss_flight_details_queries,
  })
