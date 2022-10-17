import datetime
from typing import List, Optional

from monitoring.monitorlib import fetch
from implicitdict import ImplicitDict
from monitoring.uss_qualifier.common_data_definitions import Severity


# TODO: Remove this issue structure in favor of standard report structure
class Issue(ImplicitDict):
    timestamp: Optional[str]
    """Time the issue was discovered"""

    test_code: str
    """Code corresponding to check generating this issue"""

    relevant_requirements: List[str] = []
    """Requirements that this issue relates to"""

    severity: Severity
    """How severe the issue is"""

    injection_target: str
    """Issue is related to data injected into this target.

  This is generally a Service Provider for RID.
  """

    observation_source: str
    """Issue is related to the system state observed using this target.

  This is a Display Application/Provider for RID.
  """

    subject: Optional[str]
    """Identifier of the subject of this issue, if applicable.

  This may be a flight ID, or ID of other object central to the issue.
  """

    summary: str
    """Human-readable summary of the issue"""

    details: str
    """Human-readable description of the issue"""

    queries: List[fetch.Query]
    """Description of HTTP requests relevant to this issue"""

    def __init__(self, **kwargs):
        super(Issue, self).__init__(**kwargs)
        if "timestamp" not in kwargs:
            self.timestamp = datetime.datetime.utcnow().isoformat()


class TestRunnerError(RuntimeError):
    """An error encountered when interacting with a DSS or a USS"""

    def __init__(self, msg, issue: Issue):
        super(TestRunnerError, self).__init__(msg)
        self.issue = issue


class TestStepError(RuntimeError):
    """An error encountered when interacting with a DSS or a USS"""

    def __init__(self, msg, step):
        super(TestStepError, self).__init__(msg)
        self.step = step
