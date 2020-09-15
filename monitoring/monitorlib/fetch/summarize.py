import copy
from typing import Dict

from . import rid, scd


def isas(fetched: rid.FetchedISAs) -> Dict:
  summary = {}
  if fetched.success:
    for isa_id, isa in fetched.isas.items():
      if isa.flights_url not in summary:
        summary[isa.flights_url] = {}
      isa_summary = copy.deepcopy(isa)
      if 'id' in isa_summary:
        del isa_summary['id']
      if 'owner' in isa_summary:
        del isa_summary['owner']
      isa_key = '{} ({})'.format(isa.id, isa.owner)
      summary[isa.flights_url][isa_key] = isa_summary
  else:
    summary['error'] = fetched.error
  return summary


def _entity(fetched: scd.FetchedEntities, id: str) -> Dict:
  entity = fetched.entities_by_id[id]
  if entity.success:
    return {
      'reference': {
        'dss': fetched.dss_query.references_by_id.get(id, None),
        'uss': entity.reference,
      },
      'details': entity.details,
    }
  else:
    return {
      'error': entity.error,
    }


def entities(fetched: scd.FetchedEntities) -> Dict:
  if fetched.success:
    return {
      'new': {id: _entity(fetched, id) for id in fetched.new_entities_by_id},
      'cached': {id: _entity(fetched, id) for id in fetched.cached_entities_by_id},
    }
  else:
    return {
      'error': fetched.error,
    }


def flights(fetched: rid.FetchedFlights) -> Dict:
  if fetched.success:
    isas_by_url = {}
    owners_by_url = {}
    for isa_id, isa in fetched.dss_isa_query.isas.items():
      if isa.flights_url not in isas_by_url:
        isas_by_url[isa.flights_url] = {}
      isa_info = copy.deepcopy(isa)
      del isa_info['id']
      isas_by_url[isa.flights_url][isa_id] = isa_info
      owners_by_url[isa.flights_url] = isa.owner

    summary = {}
    for url, flights_result in fetched.uss_flight_queries.items():
      if flights_result.success:
        owner = owners_by_url[url]
        isas = isas_by_url[url]
        for rid_flight in flights_result.flights:
          flight = copy.deepcopy(rid_flight)
          flight['isas'] = isas
          if rid_flight.id in fetched.uss_flight_details_queries:
            flight['details'] = fetched.uss_flight_details_queries[rid_flight.id].details
          summary['{} ({})'.format(rid_flight.id, owner)] = flight
    return summary
  else:
    return {
      'errors': fetched.errors
    }
