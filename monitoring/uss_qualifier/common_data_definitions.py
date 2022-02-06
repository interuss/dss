from enum import Enum

from monitoring.monitorlib.typing import ImplicitDict


class Severity(object):
  Critical = 'Critical'
  """The system does not function correctly on a basic level."""

  High = 'High'
  """The system may superficially function, but does not meet requirements."""

  Low = 'Low'
  """The system behaves correctly, but could be improved."""


class SubjectType(str, Enum):
    InjectedFlight = 'InjectedFlight'
    OperationalIntent = 'OperationalIntent'


class IssueSubject(ImplicitDict):
    subject_type: SubjectType
    subject: str
