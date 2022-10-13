import datetime
import enum
from typing import Dict, List, Tuple

import arrow
from termcolor import colored


class Change(enum.Enum):
    NOCHANGE = 0
    ADDED = 1
    CHANGED = 2
    REMOVED = 3

    @classmethod
    def color_of(cls, change) -> str:
        if change == Change.NOCHANGE:
            return "grey"
        elif change == Change.ADDED:
            return "green"
        elif change == Change.CHANGED:
            return "yellow"
        elif change == Change.REMOVED:
            return "red"
        raise ValueError("Invalid Change type")


def _update_overall(overall: Change, field: Change):
    if overall == Change.CHANGED:
        return Change.CHANGED
    if overall == Change.NOCHANGE:
        return field
    if overall == Change.ADDED:
        if field == Change.ADDED or field == Change.NOCHANGE:
            return Change.ADDED
        else:
            return Change.CHANGED
    if overall == Change.REMOVED:
        if field == Change.REMOVED or field == Change.NOCHANGE:
            return Change.REMOVED
        else:
            return Change.CHANGED
    raise ValueError("Unexpected change configuration")


def dict_changes(a: Dict, b: Dict) -> Tuple[Dict, Dict, Change]:
    values = {}
    changes = {}
    overall = Change.NOCHANGE

    for k, v1 in b.items():
        v0 = a.get(k, {})
        if isinstance(v1, dict):
            field_values, field_changes, change = dict_changes(v0, v1)
            if len(field_values) >= 2:
                values[k] = field_values
                changes[k] = field_changes
                changes[k]["__self__"] = change
            elif len(field_values) == 1:
                field_k = next(iter(field_values.keys()))
                k = k + "." + field_k
                values[k] = field_values[field_k]
                changes[k] = field_changes[field_k]
        else:
            if v0 == v1:
                change = Change.NOCHANGE
            else:
                values[k] = v1
                if k not in a:
                    change = Change.ADDED
                else:
                    change = Change.CHANGED
                changes[k] = change
        overall = _update_overall(overall, change)

    for k, v0 in a.items():
        if k not in b:
            if isinstance(v0, dict):
                values[k], changes[k], change = dict_changes(v0, {})
            else:
                values[k] = v0
                change = Change.REMOVED
                changes[k] = change
            overall = _update_overall(overall, change)

    return values, changes, overall


def diff_lines(values: Dict, changes: Dict) -> List[str]:
    lines = []
    for k, v in values.items():
        c = changes[k]
        if isinstance(v, dict):
            if "__self__" in c:
                lines.append(colored(k, Change.color_of(c["__self__"])) + ":")
            else:
                lines.append(k + ":")
            lines.extend("  " + line for line in diff_lines(v, c))
        else:
            if c == Change.ADDED:
                lines.append(colored("{}: {}".format(k, v), "green"))
            elif c == Change.CHANGED:
                lines.append(k + ": " + colored(str(v), "yellow"))
            elif c == Change.REMOVED:
                lines.append(colored("{}: {}".format(k, v), "red"))
    return lines


def format_timedelta(td: datetime.timedelta) -> str:
    """Produce a human-readable string describing a timedelta.
    Args:
      td: datetime.timedelta to format.
    Return:
      Formatted timedelta that looks like HH:MM:SS or XXXdHH:MM:SS where XXX is
      number of days, with or without a leading negative sign.
    """
    seconds = int(td.total_seconds())
    if seconds < 0:
        seconds = -seconds
        sign = "-"
    else:
        sign = ""
    periods = (("%d", 60 * 60 * 24), ("%02d", 60 * 60), ("%02d", 60), ("%02d", 1))
    has_days = seconds >= periods[0][1]

    segments = []
    for format_string, period_seconds in periods:
        period_value, seconds = divmod(seconds, period_seconds)
        segments.append(format_string % period_value)

    if has_days:
        return sign + "{:s}d{:s}:{:s}:{:s}".format(*segments)
    else:
        return sign + "{:s}:{:s}:{:s}".format(*segments[1:])


def make_datetime(t) -> datetime.datetime:
    if isinstance(t, str):
        return arrow.get(t).datetime
    elif isinstance(t, datetime.datetime):
        return arrow.get(t).datetime
    else:
        raise ValueError("Could not convert {} to datetime".format(str(type(t))))
