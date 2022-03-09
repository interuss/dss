from monitoring.monitorlib.clients.scd_automated_testing import QueryError
from monitoring.uss_qualifier.scd.data_interfaces import TestStep
from monitoring.uss_qualifier.scd.executor.errors import TestRunnerError
from monitoring.uss_qualifier.scd.executor.target import TestTarget
from monitoring.uss_qualifier.scd.reports import TestStepReference


def execute(self, step: TestStep, step_ref: TestStepReference, target: TestTarget) -> None:
    print("[SCD]     Step: Delete flight {} in {}".format(step.delete_flight.flight_name, target.name))
    try:
        resp, query = target.delete_flight(step.delete_flight.flight_name)
        self.report_recorder.capture_interaction(step_ref, query, 'Delete flight as part of test sequence')
    except QueryError as e:
        interaction_id = self.report_recorder.capture_interaction(step_ref, e.query, 'Delete flight as part of test sequence')
        issue = self.report_recorder.capture_deletion_unknown_issue(
            interaction_id=interaction_id,
            summary="Deletion request was unsuccessful.",
            details="Deletion attempt failed with status {}.".format(e.query.status_code),
            flight_name=step.delete_flight.flight_name,
            target_name=target.name,
            uss_role=self.get_target_role(target.name)
        )
        raise TestRunnerError("Unsuccessful attempt to delete flight {}".format(step.inject_flight.name), issue)
