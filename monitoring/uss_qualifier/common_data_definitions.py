from enum import Enum

from implicitdict import ImplicitDict


class Severity(str, Enum):
    Critical = "Critical"
    """The system does not function correctly on a basic level.
    
    Error is unrecoverable and left the system dirty preventing subsequent correct
    test run execution. For instance, inability to clean a created flight which
    will conflict with another run with a different combination of targets.
    """

    High = "High"
    """The system may superficially function, but does not meet requirements.
    
    Error interrupts a test run but likely doesn't impact subsequent test runs.
    For instance, the test failed, but all created resources can be cleaned up
    during teardown by the test driver.
    """

    Medium = "Medium"
    """The system functions, but does not meet requirements.
    
    Further test steps can likely be executed without impact.
    """

    Low = "Low"
    """The system behaves correctly, but could be improved.
    
    Further test steps can be executed without impact.
    """


class SubjectType(str, Enum):
    InjectedFlight = "InjectedFlight"
    OperationalIntent = "OperationalIntent"


class IssueSubject(ImplicitDict):
    subject_type: SubjectType
    subject: str
