from typing import Optional, List

from implicitdict import ImplicitDict

from monitoring.uss_qualifier.fileio import load_dict_with_references
from monitoring.uss_qualifier.requirements.documentation import RequirementSetID
from monitoring.uss_qualifier.resources.definitions import ResourceCollection
from monitoring.uss_qualifier.suites.definitions import (
    TestSuiteDeclaration,
    TestSuiteActionDeclaration,
)

ParticipantID = str
"""String that refers to a participant being qualified by uss_qualifier"""


class TestConfiguration(ImplicitDict):
    action: TestSuiteActionDeclaration
    """The action this test configuration wants to run (usually a test suite)"""

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


class USSQualifierConfigurationV1(ImplicitDict):
    test_run: Optional[TestConfiguration] = None
    """If specified, configuration describing how to perform a test run"""

    artifacts: Optional[ArtifactsConfiguration] = None
    """If specified, configuration describing the artifacts related to the test run"""


class USSQualifierConfiguration(ImplicitDict):
    v1: Optional[USSQualifierConfigurationV1]
    """Configuration in version 1 format"""

    @staticmethod
    def from_string(config_string: str) -> "USSQualifierConfiguration":
        return ImplicitDict.parse(
            load_dict_with_references(config_string), USSQualifierConfiguration
        )
