from monitoring.monitorlib.fetch import rid, scd, summarize
from monitoring.monitorlib import formatting


def isa_diff_text(a: rid.FetchedISAs, b: rid.FetchedISAs) -> str:
  """Create text to display to a real-time user describing a change in ISAs."""
  a_summary = summarize.isas(a) if a else {}
  b_summary = summarize.isas(b) if b else {}
  if b is not None and b.success and a is not None and not a.success:
    a_summary = {}
  if a is not None and a.success and b is not None and not b.success:
    a_summary = {}
  values, changes, _ = formatting.dict_changes(a_summary, b_summary)
  return '\n'.join(formatting.diff_lines(values, changes))


def entity_diff_text(a: scd.FetchedEntities, b: scd.FetchedEntities) -> str:
  """Create text to display to a real-time user describing a change in Entities."""
  a_summary = summarize.entities(a) if a else {}
  b_summary = summarize.entities(b) if b else {}
  if b is not None and b.success and a is not None and not a.success:
    a_summary = {}
  if a is not None and a.success and b is not None and not b.success:
    a_summary = {}
  values, changes, _ = formatting.dict_changes(a_summary, b_summary)
  return '\n'.join(formatting.diff_lines(values, changes))
