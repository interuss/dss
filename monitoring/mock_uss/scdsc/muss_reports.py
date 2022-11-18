import datetime, json
from typing import Dict, List, Optional
from implicitdict import ImplicitDict
from monitoring.monitorlib import fetch
from pathlib import Path

InteractionID = str


class MockUssTestContext(ImplicitDict):
    test_name: str
    """Name of the test"""

    test_case: str
    """Testcase describing the received request, action and the expected result"""


class Issue(ImplicitDict):
    timestamp: Optional[str]
    """Time the issue was discovered"""

    context: MockUssTestContext
    """The context describing what is being tested"""

    uss_role: str
    """Role USS was performing in the test when the issue occurred"""

    target: str
    """Issue is related to this USS/DSS"""

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


class Interaction(ImplicitDict):
    interaction_id: InteractionID
    """ID of this interaction (used to refer to this interaction from an issue)"""

    purpose: str
    """Intended purpose of the interaction - eg. to return the appropriate response"""

    context: MockUssTestContext
    """The context describing what is being tested"""

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


class MussReport(ImplicitDict):
    findings: Findings = Findings()

    def save(self):
        filepath = "./report/report_mock_uss_scdsc_messagesigning.json"
        with open(filepath, "w") as f:
            f.write(json.dumps(self, indent=4, default=str))
        print("[Mock USS] Report saved to {}".format(filepath))

    def reset(self):
        self.findings.issues = []
        self.findings.interactions = []
