from abc import ABC, abstractmethod
from datetime import datetime
from enum import Enum
import inspect
from typing import Callable, Dict, List, Optional, TypeVar, Union, Set

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
from monitoring.uss_qualifier.resources.definitions import ResourceTypeName, ResourceID


class ScenarioPhase(str, Enum):
    Undefined = "Undefined"
    NotStarted = "NotStarted"
    ReadyForTestCase = "ReadyForTestCase"
    ReadyForTestStep = "ReadyForTestStep"
    RunningTestStep = "RunningTestStep"
    ReadyForCleanup = "ReadyForCleanup"
    CleaningUp = "CleaningUp"
    Complete = "Complete"


class TestScenario(ABC):
    """Instance of a test scenario, ready to run after construction.

    Concrete subclasses of TestScenario must:
      1) Implement a constructor that accepts only parameters with types that are subclasses of Resource
      2) Call TestScenario.__init__ from the subclass's __init__
    """

    documentation: TestScenarioDocumentation
    on_failed_check: Optional[Callable[[FailedCheck], None]] = None
    _phase: ScenarioPhase = ScenarioPhase.Undefined
    _scenario_report: Optional[TestScenarioReport] = None
    _current_case: Optional[TestCaseDocumentation] = None
    _case_report: Optional[TestCaseReport] = None
    _current_step: Optional[TestStepDocumentation] = None
    _step_report: Optional[TestStepReport] = None

    def __init__(self):
        self.documentation = parse_documentation(self.__class__)
        self._phase = ScenarioPhase.NotStarted

    @abstractmethod
    def run(self):
        raise NotImplementedError(
            "A concrete test scenario must implement `run` method"
        )

    def cleanup(self):
        """Test scenarios needing to clean up after attempting to run should override this method."""
        self.skip_cleanup()

    def me(self) -> str:
        return inspection.fullname(self.__class__)

    def _make_scenario_report(self, information: Optional[str] = None) -> None:
        self._scenario_report = TestScenarioReport(
            name=self.documentation.name,
            documentation_url=self.documentation.url,
            start_time=StringBasedDateTime(datetime.utcnow()),
            cases=[],
        )
        if information is not None:
            self._scenario_report.information = information

    def _expect_phase(self, expected_phase: Union[ScenarioPhase, Set[ScenarioPhase]]):
        if isinstance(expected_phase, ScenarioPhase):
            expected_phase = {expected_phase}
        if self._phase not in expected_phase:
            caller = inspect.stack()[1].function
            acceptable_phases = ", ".join(expected_phase)
            raise RuntimeError(
                f"Test scenario `{self.me()}` was {self._phase} when {caller} was called (expected {acceptable_phases})"
            )

    def begin_test_scenario(self, information: Optional[str] = None) -> None:
        self._expect_phase(ScenarioPhase.NotStarted)
        self._make_scenario_report(information)
        self._phase = ScenarioPhase.ReadyForTestCase

    def begin_test_case(self, name: str) -> None:
        self._expect_phase(ScenarioPhase.ReadyForTestCase)
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
        self._phase = ScenarioPhase.ReadyForTestStep

    def begin_test_step(self, name: str) -> None:
        self._expect_phase(ScenarioPhase.ReadyForTestStep)
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
        self._phase = ScenarioPhase.RunningTestStep

    def record_query(self, query: fetch.Query) -> None:
        self._expect_phase({ScenarioPhase.RunningTestStep, ScenarioPhase.CleaningUp})
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
        self._expect_phase({ScenarioPhase.RunningTestStep, ScenarioPhase.CleaningUp})
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
        self._expect_phase(ScenarioPhase.RunningTestStep)
        self._step_report.end_time = StringBasedDateTime(datetime.utcnow())
        self._current_step = None
        self._step_report = None
        self._phase = ScenarioPhase.ReadyForTestStep

    def end_test_case(self) -> None:
        self._expect_phase(ScenarioPhase.ReadyForTestStep)
        self._case_report.end_time = StringBasedDateTime(datetime.utcnow())
        self._current_case = None
        self._case_report = None
        self._phase = ScenarioPhase.ReadyForTestCase

    def end_test_scenario(self) -> None:
        self._expect_phase(ScenarioPhase.ReadyForTestCase)
        self._scenario_report.end_time = StringBasedDateTime(datetime.utcnow())
        self._phase = ScenarioPhase.ReadyForCleanup

    def go_to_cleanup(self) -> None:
        self._expect_phase(
            {
                ScenarioPhase.ReadyForTestCase,
                ScenarioPhase.ReadyForTestStep,
                ScenarioPhase.RunningTestStep,
                ScenarioPhase.ReadyForCleanup,
            }
        )
        self._phase = ScenarioPhase.ReadyForCleanup

    def begin_cleanup(self) -> None:
        self._expect_phase(ScenarioPhase.ReadyForCleanup)
        if "cleanup" not in self.documentation or self.documentation.cleanup is None:
            raise RuntimeError(
                f"Test scenario `{self.me()}` attempted to begin_cleanup, but no cleanup step is documented"
            )
        self._current_step = self.documentation.cleanup
        self._step_report = TestStepReport(
            name=self._current_step.name,
            documentation_url=self._current_step.url,
            start_time=StringBasedDateTime(datetime.utcnow()),
            failed_checks=[],
        )
        self._scenario_report.cleanup = self._step_report
        self._phase = ScenarioPhase.CleaningUp

    def skip_cleanup(self) -> None:
        self._expect_phase(ScenarioPhase.ReadyForCleanup)
        if "cleanup" in self.documentation and self.documentation.cleanup is not None:
            raise RuntimeError(
                f"Test scenario `{self.me()}` skipped cleanup even though a cleanup step is documented"
            )
        self._phase = ScenarioPhase.Complete

    def end_cleanup(self) -> None:
        self._expect_phase(ScenarioPhase.CleaningUp)
        self._step_report.end_time = StringBasedDateTime(datetime.utcnow())
        self._phase = ScenarioPhase.Complete

    def record_execution_error(self, e: Exception) -> None:
        if self._phase == ScenarioPhase.Complete:
            raise RuntimeError(
                f"Test scenario `{self.me()}` indicated an execution error even though it was already Complete"
            )
        if self._scenario_report is None:
            self._make_scenario_report()
        self._scenario_report.execution_error = ErrorReport.create_from_exception(e)
        self._scenario_report.successful = False
        self._phase = ScenarioPhase.Complete

    def get_report(self) -> TestScenarioReport:
        if self._scenario_report is None:
            self._make_scenario_report()
        if "execution_error" not in self._scenario_report:
            try:
                self._expect_phase(ScenarioPhase.Complete)
            except RuntimeError as e:
                self.record_execution_error(e)

        # Evaluate success
        self._scenario_report.successful = (
            "execution_error" not in self._scenario_report
        )
        for case_report in self._scenario_report.cases:
            for step_report in case_report.steps:
                for failed_check in step_report.failed_checks:
                    if failed_check.severity != Severity.Low:
                        self._scenario_report.successful = False
        if "cleanup" in self._scenario_report:
            for failed_check in self._scenario_report.cleanup.failed_checks:
                if failed_check.severity != Severity.Low:
                    self._scenario_report.successful = False

        return self._scenario_report


TestScenarioType = TypeVar("TestScenarioType", bound=TestScenario)


class TestScenarioDeclaration(ImplicitDict):
    scenario_type: str
    """Type of test scenario, expressed as a Python class name qualified relative to this `scenarios` module"""

    resources: Dict[str, str] = {}
    """Mapping of resource parameter (additional argument to concrete test scenario constructor) to ID of resource to use"""

    def make_test_scenario(
        self, resource_pool: Dict[ResourceID, ResourceTypeName]
    ) -> TestScenario:
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
            if arg_name not in resource_pool:
                available_pool = ", ".join(resource_pool)
                raise ValueError(
                    f'Resource to populate test scenario argument "{arg_name}" was not found in the resource pool when trying to create {self.scenario_type} test scenario (resource pool: {available_pool})'
                )
            constructor_args[arg_name] = resource_pool[arg_name]

        return scenario_type(**constructor_args)
