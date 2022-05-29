import datetime, json
from enum import Enum
from typing import Dict, List, Optional

from monitoring.monitorlib import fetch
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.uss_qualifier.common_data_definitions import IssueSubject, Severity
from monitoring.uss_qualifier.scd.configuration import SCDQualifierTestConfiguration
from monitoring.uss_qualifier.scd.data_interfaces import AutomatedTestContext


InteractionID = str


class Issue(ImplicitDict):
    timestamp: Optional[str]
    """Time the issue was discovered"""

    context: AutomatedTestContext
    """Test context in which issue was discovered"""

    check_code: str
    """Code corresponding to check generating this issue"""

    uss_role: str
    """Role USS was performing in the test when the issue occurred"""

    target: str
    """Issue is related to this USS/DSS"""

    relevant_requirements: List[str] = []
    """Requirements that this issue relates to"""

    severity: Severity
    """How severe the issue is"""

    subject: Optional[IssueSubject]
    """Identifier of the subject of this issue, if applicable.

    This may be a flight ID, or ID of other object central to the issue.
    """

    summary: str
    """Human-readable summary of the issue"""

    details: str
    """Human-readable description of the issue"""

    interactions: List[InteractionID]
    """Description of interactions relevant to this issue"""

    def __init__(self, **kwargs):
        super(Issue, self).__init__(**kwargs)
        if "timestamp" not in kwargs:
            self.timestamp = datetime.datetime.utcnow().isoformat()


class TestPhase(str, Enum):
    """Phase of a test"""

    Initialization = "Initialization"
    Test = "Test"
    Cleanup = "Cleanup"


class TestStepReference(ImplicitDict):
    name: str
    """Name of test step for which this interaction was performed."""

    index: int
    """Step of test for which this interaction was performed. 0-based indexed."""

    phase: TestPhase
    """Phase in which the interaction was performed."""


class Interaction(ImplicitDict):
    interaction_id: InteractionID
    """ID of this interaction (used to refer to this interaction from an issue)"""

    purpose: str
    """Intended purpose of the interaction"""

    context: AutomatedTestContext
    """Context in which this interaction was performed."""

    test_step: TestStepReference
    """Step of test for which this interaction was performed."""

    query: fetch.Query
    """Interaction performed (flight injection, DSS query, USS query, etc)"""


class Findings(ImplicitDict):
    issues: List[Issue] = []
    interactions: List[Interaction] = []

    def add_interaction(self, interaction: Interaction):
        self.interactions.append(interaction)

    def add_issue(self, issue: Issue):
        self.issues.append(issue)

    def critical_issues(self) -> List[Issue]:
        return list(filter(lambda issue: issue.severity.Critical, self.issues))

    def __repr__(self):
        return "[{} issues in {} interactions]".format(
            len(self.issues), len(self.interactions)
        )


class TargetInformation(ImplicitDict):
    status: str
    version: str


class Report(ImplicitDict):
    qualifier_version: str
    configuration: SCDQualifierTestConfiguration
    targets_information: Dict[str, TargetInformation]
    findings: Findings = Findings()

    def save(self):
        filepath = "./report_scd.json"
        with open(filepath, "w") as f:
            json.dump(self, f)
        print("[SCD] Report saved to {}".format(filepath))
