from typing import Optional, List

from implicitdict import ImplicitDict

from monitoring.uss_qualifier.fileio import load_dict_with_references
from monitoring.uss_qualifier.requirements.documentation import RequirementSetID
from monitoring.uss_qualifier.resources.definitions import ResourceCollection
from monitoring.uss_qualifier.suites.definitions import TestSuiteDeclaration


ParticipantID = str
"""String that refers to a participant being qualified by uss_qualifier"""


class TestConfiguration(ImplicitDict):
    test_suite: TestSuiteDeclaration
    """The test suite this test configuration wants to run"""

    resources: ResourceCollection
    """Declarations for resources used by the test suite"""


class TestedRole(ImplicitDict):
    name: str
    """Name of role"""

    requirement_set: RequirementSetID
    """Set of requirements a participant must satisfy to fulfill the role"""

    participants: List[ParticipantID]
    """Participants fulfilling the role"""


class TestedRolesConfiguration(ImplicitDict):
    report_path: str
    """Path of HTML file to contain a requirements-based summary of the test report"""

    roles: List[TestedRole]
    """Roles (and participants filling those roles) tested by the test run"""


class GraphConfiguration(ImplicitDict):
    gv_path: str
    """Path of GraphViz (.gv) text file to contain a visualization of the test run"""


class ArtifactsConfiguration(ImplicitDict):
    report_path: Optional[str] = None
    """File name of the report to write (if test_config provided) or read (if test_config not provided)"""

    graph: Optional[GraphConfiguration] = None
    """If specified, configuration describing a desired graph visualization summarizing the test run"""

    tested_roles: Optional[TestedRolesConfiguration] = None
    """If specified, configuration describing a desired report summarizing tested requirements for each specified participant and role"""


class USSQualifierConfiguration(ImplicitDict):
    test_run: Optional[TestConfiguration] = None
    """If specified, configuration describing how to perform a test run"""

    artifacts: Optional[ArtifactsConfiguration] = None
    """If specified, configuration describing the artifacts related to the test run"""

    @staticmethod
    def from_string(config_string: str) -> "USSQualifierConfiguration":
        return ImplicitDict.parse(
            load_dict_with_references(config_string), USSQualifierConfiguration
        )
