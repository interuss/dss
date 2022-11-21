from implicitdict import ImplicitDict

from monitoring.uss_qualifier.fileio import load_dict_with_references
from monitoring.uss_qualifier.resources.definitions import ResourceCollection
from monitoring.uss_qualifier.suites.definitions import TestSuiteDeclaration


class TestConfiguration(ImplicitDict):
    test_suite: TestSuiteDeclaration
    """The test suite this test configuration wants to run"""

    resources: ResourceCollection
    """Declarations for resources used by the test suite"""

    @staticmethod
    def from_string(config_string: str) -> "TestConfiguration":
        return ImplicitDict.parse(
            load_dict_with_references(config_string), TestConfiguration
        )
