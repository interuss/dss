import uuid
from monitoring.monitorlib import fetch
from monitoring.mock_uss.scdsc.muss_reports import (
    Interaction,
    MussReport,
    Issue,
    InteractionID,
    MockUssTestContext,
)

class MussReportRecorder:
    """Class providing helper to capture interactions and issues in a report"""

    def __init__(self, report: MussReport):
        self.reprt = report

    def capture_interaction(
        self, query: fetch.Query, purpose: str, test_context=None
    ) -> InteractionID:
        interaction_id = str(uuid.uuid4())
        interaction = Interaction(
            interaction_id=interaction_id,
            purpose=purpose,
            query=query,
            context=test_context
        )
        self.reprt.findings.add_interaction(interaction)
        return interaction_id

    def capture_issue(self, issue: Issue):
        self.reprt.findings.add_issue(issue)
