from dataclasses import dataclass
from typing import List, Dict

from jinja2 import Environment, PackageLoader

from monitoring.uss_qualifier.reports.report import ParticipantID, TestRunReport
from monitoring.uss_qualifier.requirements.documentation import (
    RequirementSet,
    RequirementSetID,
    get_requirement_set,
)
from monitoring.uss_qualifier.scenarios.documentation.requirements import (
    TestedRequirement,
    evaluate_requirements,
)


def _all_participants(requirements: List[TestedRequirement]) -> List[ParticipantID]:
    participants = set()
    for requirement in requirements:
        for participant_id in requirement.participant_performance:
            participants.add(participant_id)
    result = list(participants)
    result.sort()
    return result


def _render_requirement_table(
    env,
    requirements: List[TestedRequirement],
    participants: List[ParticipantID],
    requirement_set_title: str,
) -> str:
    rows = [["Requirement"] + participants]
    for requirement in requirements:
        cols = [requirement.requirement_id]
        for participant in participants:
            performance = requirement.participant_performance.get(participant, None)
            if performance is None:
                cols.append("")
            else:
                n_total = len(performance.successes) + len(performance.failures)
                percentage_successful = 100 * len(performance.successes) / n_total
                cols.append("{:.0f}%".format(percentage_successful))
        rows.append(cols)

    template = env.get_template("tested_requirement_set.html")
    return template.render(rows=rows, requirement_set_title=requirement_set_title)


@dataclass
class TestedRequirementsTable(object):
    participants: List[ParticipantID]
    requirement_set: RequirementSet


def _parse_role_arguments(args: List[str]) -> List[TestedRequirementsTable]:
    tables: Dict[RequirementSetID, List[ParticipantID]] = {}
    for arg in args:
        cols = arg.split("=")
        if len(cols) != 2:
            raise ValueError(
                f'Argument "{arg}" is invalid; arguments must be in the form of <PARTICIPANT_ID>[,<PARTICIPANT_ID>,...]=<REQUIREMENT_SET_ID>'
            )
        req_set_id = cols[1]
        if req_set_id not in tables:
            tables[req_set_id] = []
        tables[req_set_id].extend(ParticipantID(s.strip()) for s in cols[0].split(","))
    return [
        TestedRequirementsTable(
            requirement_set=get_requirement_set(RequirementSetID(k)), participants=v
        )
        for k, v in tables.items()
    ]


def generate_tested_requirements(
    report: TestRunReport, role_arguments: List[str]
) -> str:
    env = Environment(loader=PackageLoader(__name__))
    tables = _parse_role_arguments(role_arguments)
    requirements = evaluate_requirements(report)
    tested_reqs_by_id = {tr.requirement_id: tr for tr in requirements}
    unclassified_reqs = set(tr.requirement_id for tr in requirements)

    content = ""
    for table in tables:
        table_reqs = []
        for r in table.requirement_set.requirement_ids:
            if r in tested_reqs_by_id:
                table_reqs.append(tested_reqs_by_id[r])
                if r in unclassified_reqs:
                    unclassified_reqs.remove(r)
            else:
                table_reqs.append(
                    TestedRequirement(requirement_id=r, participant_performance={})
                )
        content += _render_requirement_table(
            env,
            table_reqs,
            table.participants,
            table.requirement_set.name,
        )

    unclassified_tested_requirements = [
        tr for tr in requirements if tr.requirement_id in unclassified_reqs
    ]
    content += _render_requirement_table(
        env,
        unclassified_tested_requirements,
        _all_participants(unclassified_tested_requirements),
        "Unclassified requirements",
    )

    template = env.get_template("tested_requirements.html")
    return template.render(content=content)
