import uuid
from monitoring.monitorlib import fetch
from monitoring.uss_qualifier.common_data_definitions import Severity
from monitoring.uss_qualifier.scd.data_interfaces import (
    KnownIssueFields,
    FlightInjectionAttempt,
    AutomatedTestContext,
)
from monitoring.uss_qualifier.scd.reports import (
    Interaction,
    Report,
    Issue,
    InteractionID,
    TestStepReference,
)


class ReportRecorder:
    """Class providing helper to capture interactions and issues in a report"""

    def __init__(self, report: Report, context: AutomatedTestContext):
        self.report = report
        self.context = context

    def capture_interaction(
        self, step_ref: TestStepReference, query: fetch.Query, purpose: str
    ) -> InteractionID:
        interaction_id = str(uuid.uuid4())
        interaction = Interaction(
            interaction_id=interaction_id,
            purpose=purpose,
            test_step=step_ref,
            context=self.context,
            query=query,
        )
        self.report.findings.add_interaction(interaction)
        return interaction_id

    def capture_injection_issue(
        self,
        interaction_id: InteractionID,
        target_name: str,
        attempt: FlightInjectionAttempt,
        known_issue: KnownIssueFields,
    ):
        issue = Issue(
            context=self.context,
            check_code=known_issue.test_code,
            relevant_requirements=known_issue.relevant_requirements,
            severity=known_issue.severity,
            subject=known_issue.subject,
            summary=known_issue.summary,
            details=known_issue.details,
            target=target_name,
            uss_role=attempt.injection_target.uss_role,
            interactions=[interaction_id],
        )
        self.capture_issue(issue)
        return issue

    def capture_issue(self, issue: Issue):
        self.report.findings.add_issue(issue)

    def capture_injection_unknown_issue(
        self,
        interaction_id: InteractionID,
        summary: str,
        details: str,
        target_name: str,
        attempt: FlightInjectionAttempt,
    ):
        issue = Issue(
            context=self.context,
            check_code="unknown",
            relevant_requirements=[],
            severity=Severity.Critical,
            subject="Unknown issue during injection attempt",
            summary=summary,
            details=details,
            target=target_name,
            uss_role=attempt.injection_target.uss_role,
            interactions=[interaction_id],
        )
        self.capture_issue(issue)
        return issue

    def capture_deletion_unknown_issue(
        self,
        interaction_id: InteractionID,
        summary: str,
        details: str,
        flight_name: str,
        target_name: str,
        uss_role: str,
    ):
        issue = Issue(
            context=self.context,
            check_code="unknown",
            relevant_requirements=[],
            severity=Severity.Critical,
            subject="Unknown issue during deletion of flight {}".format(flight_name),
            summary=summary,
            details=details,
            target=target_name,
            uss_role=uss_role,
            interactions=[interaction_id],
        )
        self.report.findings.add_issue(issue)
        return issue
