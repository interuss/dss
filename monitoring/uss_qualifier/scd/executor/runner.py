import json
import typing
import uuid
from typing import Dict

from monitoring.monitorlib import fetch
from monitoring.monitorlib.clients.scd import OperationError
from monitoring.monitorlib.clients.scd_automated_testing import QueryError
from monitoring.monitorlib.locality import Locality
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import InjectFlightResponse, DeleteFlightResponse
from monitoring.uss_qualifier.common_data_definitions import Severity
from monitoring.uss_qualifier.scd.configuration import SCDQualifierTestConfiguration
from monitoring.uss_qualifier.scd.data_interfaces import AutomatedTest, TestStep, FlightInjectionAttempt, \
    KnownIssueFields, FlightDeletionAttempt
from monitoring.uss_qualifier.scd.executor.target import TestTarget
from monitoring.uss_qualifier.scd.reports import Report, Interaction, Issue, AutomatedTestContext, TestStepTeardownIndex

class TestRunnerError(RuntimeError):
    """An error encountered when interacting with a DSS or a USS"""
    def __init__(self, msg, issue: Issue):
        super(TestRunnerError, self).__init__(msg)
        self.issue = issue


# TODO: Replace print by logging
class TestRunner:
    """A class to run automated test steps for a specific combination of targets per uss role"""

    def __init__(self, context: AutomatedTestContext, automated_test: AutomatedTest, targets: Dict[str, TestTarget], report: Report):
        self.context = context
        self.automated_test = automated_test
        self.targets = targets
        self.report = report

    def get_scd_configuration(self) -> SCDQualifierTestConfiguration:
        return SCDQualifierTestConfiguration(injection_targets=list(map(lambda t: t.config, self.targets.values())))

    def run_automated_test(self):
        for i, step in enumerate(self.automated_test.steps):
            print('[SCD]   Running step {}: {}'.format(i, step.name))
            self.execute_step(step, i)

    def teardown(self):
        """Delete resources created by this test runner."""
        print("[SCD]   Teardown {}".format(self.automated_test.name))

        def capture_teardown_interaction(query: fetch.Query):
            return self.capture_interaction(TestStepTeardownIndex, query)

        for role, target in self.targets.items():
            # TODO: Capture errors
            target.delete_all_flights(capture_teardown_interaction)

    def execute_step(self, step: TestStep, step_index: int):
        target = self.get_target(step)
        if target is None:
            self.print_targets_state()
            raise RuntimeError("[SCD] Error: Unable to identify the target managing flight {}".format(
                step.inject_flight.name if 'inject_flight' in step else step.delete_flight.flight_name
            ))

        if 'inject_flight' in step:
            print("[SCD]     Step: Inject flight {} to {}".format(step.inject_flight.name, target.name))
            try:
                resp, query = target.inject_flight(step.inject_flight)
                interaction_id = self.capture_interaction(step_index, query)
                self.evaluate_inject_flight_response(step.inject_flight, resp, interaction_id)
            except QueryError as e:
                interaction_id = self.capture_interaction(step_index, e.query)
                issue = self.capture_injection_unknown_issue(
                    "Injection request was unsuccessful",
                    "Injection attempt failed with status {}.".format(e.query.status_code),
                    step.inject_flight,
                    interaction_id
                )
                raise TestRunnerError("Unsuccessful attempt to inject flight {}".format(step.inject_flight.name), issue)

        elif 'delete_flight' in step:
            print("[SCD]     Step: Delete flight {} to {}".format(step.delete_flight.flight_name, target.name))
            # TODO: Capture errors
            resp, query = target.delete_flight(step.delete_flight.flight_name)
        else:
            raise RuntimeError("[SCD] Error: Unable to identify the action to execute for step {}".format(step.name))

    def get_managing_target(self, flight_name: str) -> typing.Optional[TestTarget]:
        """Returns the managing target which created a flight"""
        for role, target in self.targets.items():
            if target.is_managing_flight(flight_name):
                return target
        return None

    def get_target(self, step: TestStep) -> typing.Optional[TestTarget]:
        """Returns the target which should be called in the TestStep"""
        if 'inject_flight' in step:
            return self.targets[step.inject_flight.injection_target.uss_role]
        elif 'delete_flight' in step:
            return self.get_managing_target(step.delete_flight.flight_name)
        else:
            raise NotImplementedError("Unsupported step. A Test Step shall contain either a inject_flight or a delete_flight object.")

    def capture_interaction(self, step_index: int, query: fetch.Query) -> str:
        interaction_id = str(uuid.uuid4())
        interaction = Interaction(
                interaction_id=interaction_id,
                test_step=step_index,
                context=self.context,
                query=query
            )
        self.report.findings.add_interaction(interaction)
        return interaction_id

    def capture_injection_issue(self, attempt: FlightInjectionAttempt, known_issue: KnownIssueFields, interaction_id: str):
        issue = Issue(
                context=self.context,
                check_code=known_issue.test_code,
                relevant_requirements=known_issue.relevant_requirements,
                severity=known_issue.severity,
                subject=known_issue.subject,
                summary=known_issue.summary,
                details=known_issue.details,
                target=attempt.injection_target,
                uss_role=attempt.injection_target.uss_role,
                interactions=[interaction_id]
            )
        self.report.findings.add_issue(issue)
        return issue

    def capture_injection_unknown_issue(self, summary: str, details: str, attempt: FlightInjectionAttempt, interaction_id: str):
        issue = Issue(
                context=self.context,
                check_code="unknown",
                relevant_requirements=[],
                severity=Severity.Critical,
                subject="Unknown issue",
                summary=summary,
                details=details,
                target=attempt.injection_target,
                uss_role=attempt.injection_target.uss_role,
                interactions=[interaction_id]
            )
        self.report.findings.add_issue(issue)
        return issue

    def evaluate_inject_flight_response(self, attempt: FlightInjectionAttempt, resp: InjectFlightResponse, interaction_id: str) -> typing.Optional[Issue]:
        if resp.result not in attempt.known_responses.acceptable_results:
            print("[SCD]     Result: ERROR. Received {}, expected one of {}. Reason: {}".format(
                resp.result,
                attempt.known_responses.acceptable_results,
                resp.get('notes', None))
            )
            known_issue = attempt.known_responses.incorrect_result_details.get(resp.result, None)
            if known_issue:
                issue = self.capture_injection_issue(attempt, known_issue, interaction_id)
                if known_issue.severity != Severity.Low:
                    raise TestRunnerError("Failed attempt to inject flight {}: {}".format(attempt.name, known_issue.summary), issue)
            else:
                issue = self.capture_injection_unknown_issue(
                    "Injection request was unsuccessful",
                    "Injection attempt failed with unknown response {}".format(resp.result),
                    attempt,
                    interaction_id
                )
                raise TestRunnerError("Unsuccessful attempt to inject flight {}".format(attempt.name), issue)

        print("[SCD]     Result: COMPLETED")
        return None


    def print_targets_state(self):
        print("[SCD] Targets States:")
        for name, target in self.targets.items():
            print(f"[SCD]   - {name}: {target.created_flight_ids}")
