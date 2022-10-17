import inspect
import os
from typing import List, Optional, Type

from implicitdict import ImplicitDict
import marko
import marko.element
import marko.inline

from monitoring.monitorlib.inspection import fullname


RESOURCES_HEADING = "resources"
TEST_SCENARIO_SUFFIX = " test scenario"
TEST_CASE_SUFFIX = " test case"
TEST_STEP_SUFFIX = " test step"
TEST_CHECK_SUFFIX = " check"


class TestCheckDocumentation(ImplicitDict):
    name: str
    url: Optional[str] = None
    applicable_requirements: List[str]


class TestStepDocumentation(ImplicitDict):
    name: str
    url: Optional[str] = None
    checks: List[TestCheckDocumentation]


class TestCaseDocumentation(ImplicitDict):
    name: str
    url: Optional[str] = None
    steps: List[TestStepDocumentation]


class TestScenarioDocumentation(ImplicitDict):
    name: str
    url: Optional[str] = None
    resources: Optional[List[str]]
    cases: List[TestCaseDocumentation]


def _text_of(value) -> str:
    if isinstance(value, str):
        return value
    elif isinstance(value, marko.block.BlockElement):
        result = ""
        for child in value.children:
            result += _text_of(child)
        return result
    elif isinstance(value, marko.inline.InlineElement):
        if isinstance(value.children, str):
            return value.children
        result = ""
        for child in value.children:
            result += _text_of(child)
        return result
    else:
        raise NotImplementedError(
            "Cannot yet extract raw text from {}".format(value.__class__.__name__)
        )


def _length_of_section(values, start_of_section: int) -> int:
    level = values[start_of_section].level
    c = start_of_section + 1
    while c < len(values):
        if isinstance(values[c], marko.block.Heading) and values[c].level == level:
            break
        c += 1
    return c - start_of_section - 1


def _parse_test_check(values) -> TestCheckDocumentation:
    name = _text_of(values[0])[0 : -len(TEST_CHECK_SUFFIX)]

    reqs: List[str] = []
    c = 1
    while c < len(values):
        if isinstance(values[c], marko.block.Paragraph):
            for child in values[c].children:
                if isinstance(child, marko.inline.StrongEmphasis):
                    reqs.append(_text_of(child))
        c += 1

    return TestCheckDocumentation(name=name, applicable_requirements=reqs)


def _parse_test_step(values) -> TestStepDocumentation:
    name = _text_of(values[0])[0 : -len(TEST_STEP_SUFFIX)]

    checks: List[TestCheckDocumentation] = []
    c = 1
    while c < len(values):
        if isinstance(values[c], marko.block.Heading) and _text_of(
            values[c]
        ).lower().endswith(TEST_CHECK_SUFFIX):
            # Start of a test step section
            dc = _length_of_section(values, c)
            check = _parse_test_check(values[c : c + dc + 1])
            checks.append(check)
            c += dc
        else:
            c += 1

    return TestStepDocumentation(name=name, checks=checks)


def _parse_test_case(values) -> TestCaseDocumentation:
    name = _text_of(values[0])[0 : -len(TEST_CASE_SUFFIX)]

    steps: List[TestStepDocumentation] = []
    c = 1
    while c < len(values):
        if isinstance(values[c], marko.block.Heading) and _text_of(
            values[c]
        ).lower().endswith(TEST_STEP_SUFFIX):
            # Start of a test step section
            dc = _length_of_section(values, c)
            step = _parse_test_step(values[c : c + dc + 1])
            steps.append(step)
            c += dc
        else:
            c += 1

    return TestCaseDocumentation(name=name, steps=steps)


def _parse_resources(values) -> List[str]:
    resource_level = values[0].level + 1
    resources: List[str] = []
    c = 1
    while c < len(values):
        if (
            isinstance(values[c], marko.block.Heading)
            and values[c].level == resource_level
        ):
            # This is a resource
            resources.append(_text_of(values[c]))
        c += 1
    return resources


def parse_documentation(scenario: Type) -> TestScenarioDocumentation:
    # Load the .md file matching the Python file where this scenario type is defined
    doc_filename = os.path.splitext(inspect.getfile(scenario))[0] + ".md"
    if not os.path.exists(doc_filename):
        raise ValueError(
            "Test scenario `{}` does not have the required documentation file `{}`".format(
                fullname(scenario), doc_filename
            )
        )
    with open(doc_filename, "r") as f:
        doc = marko.parse(f.read())

    # Extract the scenario name from the first top-level header
    if (
        not isinstance(doc.children[0], marko.block.Heading)
        or doc.children[0].level != 1
        or not _text_of(doc.children[0]).lower().endswith(TEST_SCENARIO_SUFFIX)
    ):
        raise ValueError(
            'The first line of {} must be a level-1 heading with the name of the scenario + "{}" (e.g., "# ASTM NetRID nominal behavior test scenario")'.format(
                doc_filename, TEST_SCENARIO_SUFFIX
            )
        )
    scenario_name = _text_of(doc.children[0])[0 : -len(TEST_SCENARIO_SUFFIX)]

    # Step through the document to extract important structured components
    test_cases: List[TestCaseDocumentation] = []
    resources = None
    c = 1
    while c < len(doc.children):
        if (
            isinstance(doc.children[c], marko.block.Heading)
            and _text_of(doc.children[c]).lower().strip() == RESOURCES_HEADING
        ):
            # Start of the Resources section
            if resources is not None:
                raise ValueError(
                    'Only one major section may be titled "{}"'.format(
                        RESOURCES_HEADING
                    )
                )
            dc = _length_of_section(doc.children, c)
            resources = _parse_resources(doc.children[c : c + dc + 1])
            c += dc
        if isinstance(doc.children[c], marko.block.Heading) and _text_of(
            doc.children[c]
        ).lower().endswith(TEST_CASE_SUFFIX):
            # Start of a test case section
            dc = _length_of_section(doc.children, c)
            test_case = _parse_test_case(doc.children[c : c + dc + 1])
            test_cases.append(test_case)
            c += dc
        else:
            c += 1

    return TestScenarioDocumentation(
        # TODO: Populate the documentation URLs
        name=scenario_name,
        cases=test_cases,
        resources=resources,
        url="",
    )
