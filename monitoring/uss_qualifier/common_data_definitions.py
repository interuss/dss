from enum import Enum


class Severity(str, Enum):
  Critical = 'Critical'
  """The system does not function correctly on a basic level."""

  High = 'High'
  """The system may superficially function, but does not meet requirements."""

  Low = 'Low'
  """The system behaves correctly, but could be improved."""
  