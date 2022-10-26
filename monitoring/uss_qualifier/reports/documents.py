from typing import List

from jinja2 import Environment, PackageLoader

from monitoring.uss_qualifier.reports.report import ParticipantID
from monitoring.uss_qualifier.scenarios.documentation.requirements import Requirement


def _all_participants(requirements: List[Requirement]) -> List[ParticipantID]:
    participants = set()
    for requirement in requirements:
        for participant_id in requirement.participant_performance:
            participants.add(participant_id)
    result = list(participants)
    result.sort()
    return result


def render_requirement_table(requirements: List[Requirement]) -> str:
    all_participants = _all_participants(requirements)
    rows = [["Requirement"] + all_participants]
    for requirement in requirements:
        cols = [requirement.requirement_id]
        for participant in all_participants:
            performance = requirement.participant_performance.get(participant, None)
            if performance is None:
                cols.append("")
            else:
                n_total = len(performance.successes) + len(performance.failures)
                percentage_successful = 100 * len(performance.successes) / n_total
                cols.append("{:.0f}%".format(percentage_successful))
        rows.append(cols)

    env = Environment(loader=PackageLoader(__name__))
    template = env.get_template("tested_requirements.html")
    return template.render(rows=rows)
