from abc import ABC, abstractmethod
from datetime import datetime
import inspect
from typing import Callable, Dict, List, Optional, TypeVar

from implicitdict import ImplicitDict, StringBasedDateTime

from monitoring.monitorlib import fetch, inspection
from monitoring.uss_qualifier import scenarios as scenarios_module
from monitoring.uss_qualifier.common_data_definitions import Severity
from monitoring.uss_qualifier.reports import (
    TestScenarioReport,
    TestCaseReport,
    TestStepReport,
    FailedCheck,
    ErrorReport,
)
from monitoring.uss_qualifier.scenarios.documentation import (
    TestScenarioDocumentation,
    TestCaseDocumentation,
    TestStepDocumentation,
    parse_documentation,
)
from monitoring.uss_qualifier.resources import Resource


class TestScenario(ABC):
    """Instance of a test scenario, ready to run after construction.

    Concrete subclasses of TestScenario must:
      1) Implement a constructor that accepts only parameters with types that are subclasses of Resource
      2) Call TestScenario.__init__ from the subclass's __init__
    """

    documentation: TestScenarioDocumentation
    on_failed_check: Optional[Callable[[FailedCheck], None]] = None
    _scenario_report: Optional[TestScenarioReport] = None
    _current_case: Optional[TestCaseDocumentation] = None
    _case_report: Optional[TestCaseReport] = None
    _current_step: Optional[TestStepDocumentation] = None
    _step_report: Optional[TestStepReport] = None

    def __init__(self):
        self.documentation = parse_documentation(self.__class__)

    @abstractmethod
    def run(self):
        raise NotImplementedError(
            "A concrete test scenario must implement `run` method"
        )

    def me(self) -> str:
        return inspection.fullname(self.__class__)

    def _make_scenario_report(self) -> None:
        self._scenario_report = TestScenarioReport(
            name=self.documentation.name,
            documentation_url=self.documentation.url,
            start_time=StringBasedDateTime(datetime.utcnow()),
            cases=[],
        )

    def begin_test_scenario(self) -> None:
        if self._scenario_report is not None:
            raise RuntimeError(
                f"Test scenario `{self.me()}` had already begun when begin_test_scenario was called"
            )
        self._make_scenario_report()

    def begin_test_case(self, name: str) -> None:
        if self._scenario_report is None:
            raise RuntimeError(
                f"Test scenario `{self.me()}` had not yet begun when begin_test_case was called"
            )
        if self._case_report is not None:
            raise RuntimeError(
                f'Test case "{name}" had already begun when begin_test_case was called in test scenario `{self.me()}`'
            )
        available_cases = {c.name: c for c in self.documentation.cases}
        if name not in available_cases:
            case_list = ", ".join(available_cases)
            raise RuntimeError(
                f'Test scenario `{self.me()}` was instructed to begin_test_case "{name}", but that test case is not declared in documentation; declared cases are: {case_list}'
            )
        if name in [c.name for c in self._scenario_report.cases]:
            raise RuntimeError(
                f"Test case {name} had already run in `{self.me()}` when begin_test_case was called"
            )
        self._current_case = available_cases[name]
        self._case_report = TestCaseReport(
            name=self._current_case.name,
            documentation_url=self._current_case.url,
            start_time=StringBasedDateTime(datetime.utcnow()),
            steps=[],
        )
        self._scenario_report.cases.append(self._case_report)

    def begin_test_step(self, name: str) -> None:
        if self._scenario_report is None:
            raise RuntimeError(
                f"Test scenario `{self.me()}` had not yet begun when begin_test_step was called"
            )
        if self._case_report is None:
            raise RuntimeError(
                f"Test case for test scenario `{self.me()}` had not yet begun when begin_test_step was called"
            )
        available_steps = {c.name: c for c in self._current_case.steps}
        if name not in available_steps:
            step_list = ", ".join(available_steps)
            raise RuntimeError(
                f'Test scenario `{self.me()}` was instructed to begin_test_step "{name}" during test case "{self._current_case.name}", but that test step is not declared in documentation; declared steps are: {step_list}'
            )
        self._current_step = available_steps[name]
        self._step_report = TestStepReport(
            name=self._current_step.name,
            documentation_url=self._current_step.url,
            start_time=StringBasedDateTime(datetime.utcnow()),
            failed_checks=[],
        )
        self._case_report.steps.append(self._step_report)

    def record_query(self, query: fetch.Query) -> None:
        if self._scenario_report is None:
            raise RuntimeError(
                f"Test scenario `{self.me()}` had not yet begun when record_query was called"
            )
        if self._case_report is None:
            raise RuntimeError(
                f"Test case for test scenario `{self.me()}` had not yet begun when record_query was called"
            )
        if self._step_report is None:
            raise RuntimeError(
                f"Test step for test case {self._current_case.name} for test scenario `{self.me()}` had not yet begun when record_query was called"
            )
        if "queries" not in self._step_report:
            self._step_report.queries = []
        self._step_report.queries.append(query)

    def record_failed_check(
        self,
        name: str,
        summary: str,
        severity: Severity,
        relevant_participants: List[str],
        details: str = "",
        query_timestamps: Optional[List[datetime]] = None,
        additional_data: Optional[dict] = None,
    ) -> None:
        if self._scenario_report is None:
            raise RuntimeError(
                f"Test scenario `{self.me()}` had not yet begun when record_failed_check was called"
            )
        if self._case_report is None:
            raise RuntimeError(
                f"Test case for test scenario `{self.me()}` had not yet begun when record_failed_check was called"
            )
        if self._step_report is None:
            raise RuntimeError(
                f"Test step for test case {self._current_case.name} for test scenario `{self.me()}` had not yet begun when record_failed_check was called"
            )
        available_checks = {c.name: c for c in self._current_step.checks}
        if name not in available_checks:
            check_list = ", ".join(available_checks)
            raise RuntimeError(
                f'Test scenario `{self.me()}` was instructed to record_failed_check "{name}" during test step "{self._current_step.name}" during test case "{self._current_case.name}", but that check is not declared in documentation; declared checks are: {check_list}'
            )
        check = available_checks[name]

        kwargs = {
            "name": check.name,
            "documentation_url": check.url,
            "timestamp": StringBasedDateTime(datetime.utcnow()),
            "summary": summary,
            "details": details,
            "relevant_requirements": check.applicable_requirements,
            "severity": severity,
            "relevant_participants": relevant_participants,
        }
        if additional_data is not None:
            kwargs["additional_data"] = additional_data
        if query_timestamps is not None:
            kwargs["query_report_timestamps"] = [
                StringBasedDateTime(t) for t in query_timestamps
            ]
        failed_check = FailedCheck(**kwargs)
        self._step_report.failed_checks.append(failed_check)
        if self.on_failed_check is not None:
            self.on_failed_check(failed_check)

    def end_test_step(self) -> None:
        if self._scenario_report is None:
            raise RuntimeError(
                f"Test scenario `{self.me()}` had not yet begun when end_test_step was called"
            )
        if self._case_report is None:
            raise RuntimeError(
                f"Test case for test scenario `{self.me()}` had not yet begun when end_test_step was called"
            )
        if self._step_report is None:
            raise RuntimeError(
                f"Test step for test case {self._current_case.name} for test scenario `{self.me()}` had not yet begun when end_test_step was called"
            )
        self._step_report.end_time = StringBasedDateTime(datetime.utcnow())
        self._current_step = None
        self._step_report = None

    def end_test_case(self) -> None:
        if self._step_report is not None:
            raise RuntimeError(
                f'Test step "{self._current_step.name}" in test case "{self._current_case.name}" in test scenario `{self.me()}` had not ended when end_test_case was called'
            )
        if self._scenario_report is None:
            raise RuntimeError(
                f"Test scenario `{self.me()}` had not yet begun when end_test_case was called"
            )
        if self._case_report is None:
            raise RuntimeError(
                f"There was no active test case when end_test_case was called in test scenario `{self.me()}`"
            )
        self._case_report.end_time = StringBasedDateTime(datetime.utcnow())
        self._current_case = None
        self._case_report = None

    def end_test_scenario(self) -> None:
        if self._case_report is not None:
            raise RuntimeError(
                f'Test case "{self._current_case.name}" in test scenario `{self.me()}` had not ended when end_test_scenario was called'
            )
        if self._scenario_report is None:
            raise RuntimeError(
                f"Test scenario `{self.me()}` had not yet begun when end_test_scenario was called"
            )
        if "end_time" in self._scenario_report:
            raise RuntimeError(
                f"Test scenario `{self.me()}` had already ended when end_test_scenario was called"
            )
        self._scenario_report.end_time = StringBasedDateTime(datetime.utcnow())
        self._scenario_report.successful = True

        # Look to see if any issues were found
        for case_report in self._scenario_report.cases:
            for step_report in case_report.steps:
                for failed_check in step_report.failed_checks:
                    if failed_check.severity != Severity.Low:
                        self._scenario_report.successful = False

    def record_execution_error(self, e: Exception) -> None:
        if self._scenario_report is None:
            self._make_scenario_report()
        self._scenario_report.execution_error = ErrorReport.create_from_exception(e)
        self._scenario_report.successful = False

    def get_report(self) -> TestScenarioReport:
        if self._scenario_report is None:
            scenario_not_started = True
            self._make_scenario_report()
        else:
            scenario_not_started = False
        if "execution_error" not in self._scenario_report:
            try:
                if scenario_not_started:
                    raise RuntimeError(
                        f"Report was requested for test scenario {self.me()}, but it had not even been started"
                    )
                if self._step_report is not None:
                    raise RuntimeError(
                        f"Test step {self._current_step.name} in test case {self._current_case.name} for test scenario {self.me()} did not complete"
                    )
                if self._case_report is not None:
                    raise RuntimeError(
                        f'Test case "{self._current_case.name}" for test scenario {self.me()} did not complete'
                    )
            except RuntimeError as e:
                self.record_execution_error(e)
        return self._scenario_report


TestScenarioType = TypeVar("TestScenarioType", bound=TestScenario)


class TestScenarioDeclaration(ImplicitDict):
    scenario_type: str
    """Type of test scenario, expressed as a Python class name qualified relative to this `scenarios` module"""

    resources: Dict[str, str] = {}
    """Mapping of resource parameter (additional argument to concrete test scenario constructor) to ID of resource to use"""

    def make_test_scenario(self, resource_pool: Dict[str, Resource]) -> TestScenario:
        inspection.import_submodules(scenarios_module)
        scenario_type = inspection.get_module_object_by_name(
            scenarios_module, self.scenario_type
        )
        if not issubclass(scenario_type, TestScenario):
            raise NotImplementedError(
                "Scenario type {} is not a subclass of the TestScenario base class".format(
                    scenario_type.__name__
                )
            )

        constructor_signature = inspect.signature(scenario_type.__init__)
        constructor_args = {}
        for arg_name, arg in constructor_signature.parameters.items():
            if arg_name == "self":
                continue
            if arg_name not in self.resources:
                raise ValueError(
                    'Test scenario declaration for {} is missing a source for resource "{}"'.format(
                        self.scenario_type, arg
                    )
                )
            if self.resources[arg_name] not in resource_pool:
                raise ValueError(
                    'Resource "{}" was not found in the resource pool when trying to create test scenario "{}" ({})'.format(
                        self.resources[arg_name], self.name, self.type
                    )
                )
            constructor_args[arg_name] = resource_pool[self.resources[arg_name]]

        return scenario_type(**constructor_args)
