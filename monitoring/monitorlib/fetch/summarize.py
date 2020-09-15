import copy
from typing import Dict

from . import rid, scd


def isas(fetched: rid.FetchedISAs) -> Dict:
  summary = {'interactions': fetched}
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
      'interactions': fetched,
    }
  else:
    return {
      'error': fetched.error,
      'interactions': fetched,
    }
