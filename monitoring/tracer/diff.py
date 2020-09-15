from monitoring.monitorlib.fetch import rid, scd, summarize
from monitoring.monitorlib import formatting


def isa_diff_text(a: rid.FetchedISAs, b: rid.FetchedISAs) -> str:
  """Create text to display to a real-time user describing a change in ISAs."""
  a_summary = summarize.isas(a) if a else {}
  a_summary = summarize.limit_long_arrays(a_summary, 6)
  b_summary = summarize.isas(b) if b else {}
  b_summary = summarize.limit_long_arrays(b_summary, 6)
  if b is not None and b.success and a is not None and not a.success:
    a_summary = {}
  if a is not None and a.success and b is not None and not b.success:
    a_summary = {}
  values, changes, _ = formatting.dict_changes(a_summary, b_summary)
  return '\n'.join(formatting.diff_lines(values, changes))


def entity_diff_text(a: scd.FetchedEntities, b: scd.FetchedEntities) -> str:
  """Create text to display to a real-time user describing a change in Entities."""
  entity_type = b.dss_query.entity_type if b else (a.dss_query.entity_type if a else None)
  if entity_type and '_' in entity_type:
    entity_type = entity_type[0:entity_type.index('_')]
  a_summary = summarize.entities(a, entity_type) if a else {}
  a_summary = summarize.limit_long_arrays(a_summary, 6)
  b_summary = summarize.entities(b, entity_type) if b else {}
  b_summary = summarize.limit_long_arrays(b_summary, 6)
  if b is not None and b.success and a is not None and not a.success:
    a_summary = {}
  if a is not None and a.success and b is not None and not b.success:
    a_summary = {}
  values, changes, _ = formatting.dict_changes(a_summary, b_summary)
  return '\n'.join(formatting.diff_lines(values, changes))
