from typing import List, Dict, Union

from implicitdict import ImplicitDict

from monitoring.monitorlib.inspection import (
    import_submodules,
)
from monitoring.uss_qualifier import scenarios as scenarios_module
from monitoring.uss_qualifier.reports.report import (
    TestRunReport,
    TestSuiteReport,
    ActionGeneratorReport,
    TestSuiteActionReport,
    TestScenarioReport,
    RequirementID,
    ParticipantID,
    TestCaseReport,
    TestStepReport,
    PassedCheck,
    FailedCheck,
)
from monitoring.uss_qualifier.scenarios.documentation.definitions import (
    TestScenarioDocumentation,
    TestCaseDocumentation,
    TestStepDocumentation,
)
from monitoring.uss_qualifier.scenarios.documentation.parsing import (
    get_documentation_by_name,
)
from monitoring.uss_qualifier.suites.definitions import ActionType

JSONPath = str


class ParticipantRequirementPerformance(ImplicitDict):
    successes: List[JSONPath]
    """List of passed checks involving the requirement"""

    failures: List[JSONPath]
    """List of failed checks involving the requirement"""


class TestedRequirement(ImplicitDict):
    requirement_id: RequirementID
    """Identity of the requirement"""

    participant_performance: Dict[ParticipantID, ParticipantRequirementPerformance]
    """The performance of each involved participant on the requirement"""


def _add_check(
    check: Union[PassedCheck, FailedCheck],
    path: JSONPath,
    scenario_docs: TestScenarioDocumentation,
    case_docs: TestCaseDocumentation,
    step_docs: TestStepDocumentation,
    requirements: Dict[RequirementID, TestedRequirement],
):
    if not check.requirements:
        # Generate an implied requirement ID
        tested_requirements = [
            f"{scenario_docs.name.title()}.{case_docs.name.title()}.{step_docs.name.title()}.{check.name.title()}".replace(
                " ", ""
            )
        ]
    else:
        tested_requirements = check.requirements
    for requirement_id in tested_requirements:
        if requirement_id not in requirements:
            requirements[requirement_id] = TestedRequirement(
                requirement_id=requirement_id, participant_performance={}
            )
        requirement = requirements[requirement_id]

        participants = check.participants if check.participants else [""]
        for participant_id in participants:
            if participant_id not in requirement.participant_performance:
                requirement.participant_performance[
                    participant_id
                ] = ParticipantRequirementPerformance(successes=[], failures=[])
            performance = requirement.participant_performance[participant_id]
            if isinstance(check, PassedCheck):
                performance.successes.append(path)
            elif isinstance(check, FailedCheck):
                performance.failures.append(path)
            else:
                raise ValueError("Provided check was not a PassedCheck or FailedCheck")


def _evaluate_requirements_in_step(
    report: TestStepReport,
    scenario_docs: TestScenarioDocumentation,
    case_docs: TestCaseDocumentation,
    path: JSONPath,
    requirements: Dict[RequirementID, TestedRequirement],
) -> None:
    step_docs = case_docs.get_step_by_name(report.name)
    for i, check in enumerate(report.passed_checks):
        _add_check(
            check,
            path + f".passed_checks[{i}]",
            scenario_docs,
            case_docs,
            step_docs,
            requirements,
        )
    for i, check in enumerate(report.failed_checks):
        _add_check(
            check,
            path + f".failed_checks[{i}]",
            scenario_docs,
            case_docs,
            step_docs,
            requirements,
        )


def _evaluate_requirements_in_case(
    report: TestCaseReport,
    scenario_docs: TestScenarioDocumentation,
    path: JSONPath,
    requirements: Dict[RequirementID, TestedRequirement],
) -> None:
    case_docs = scenario_docs.get_case_by_name(report.name)
    for i, step in enumerate(report.steps):
        _evaluate_requirements_in_step(
            step, scenario_docs, case_docs, path + f".steps[{i}]", requirements
        )


def _evaluate_requirements_in_scenario(
    report: TestScenarioReport,
    path: JSONPath,
    requirements: Dict[RequirementID, TestedRequirement],
) -> None:
    scenario_docs = get_documentation_by_name(report.scenario_type)
    for i, case in enumerate(report.cases):
        _evaluate_requirements_in_case(
            case, scenario_docs, path + f".cases[{i}]", requirements
        )


def _evaluate_requirements_in_action(
    report: TestSuiteActionReport,
    path: JSONPath,
    requirements: Dict[RequirementID, TestedRequirement],
) -> None:
    if "test_suite" in report:
        _evaluate_requirements_in_suite(
            report.test_suite, path + ".test_suite", requirements
        )
    elif "action_generator" in report:
        _evaluate_requirements_in_generator(
            report.action_generator, path + ".action_generator", requirements
        )
    elif "test_scenario" in report:
        _evaluate_requirements_in_scenario(
            report.test_scenario, path + ".test_scenario", requirements
        )
    else:
        raise ValueError("Unsupported action type found in TestSuiteActionReport")


def _evaluate_requirements_in_generator(
    report: ActionGeneratorReport,
    path: JSONPath,
    requirements: Dict[RequirementID, TestedRequirement],
) -> None:
    for i, action in enumerate(report.actions):
        _evaluate_requirements_in_action(action, path + f".action[{i}]", requirements)


def _evaluate_requirements_in_suite(
    report: TestSuiteReport,
    path: JSONPath,
    requirements: Dict[RequirementID, TestedRequirement],
) -> None:
    for i, action in enumerate(report.actions):
        _evaluate_requirements_in_action(action, path + f".action[{i}]", requirements)


def evaluate_requirements(report: TestRunReport) -> List[TestedRequirement]:
    import_submodules(scenarios_module)
    reqs = {}
    _evaluate_requirements_in_action(report.report, "$.report", reqs)
    sorted_ids = list(reqs.keys())
    sorted_ids.sort()
    return [reqs[k] for k in sorted_ids]
