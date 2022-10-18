import json
import os

from implicitdict import ImplicitDict
import requests

from monitoring.uss_qualifier.resources import ResourceCollection
from monitoring.uss_qualifier.suites.definitions import TestSuiteDeclaration


class TestConfiguration(ImplicitDict):
    test_suite: TestSuiteDeclaration
    """The test suite this test configuration wants to run"""

    resources: ResourceCollection
    """Declarations for resources used by the test suite"""

    @staticmethod
    def from_string(config_string: str) -> "TestConfiguration":
        if config_string.startswith("http://") or config_string.startswith("https://"):
            # Spec is a URL
            resp = requests.get(config_string)
            if "application/json" in resp.headers.get("Content-Type", ""):
                config_dict = resp.json()
            else:
                config_dict = json.loads(resp.content.decode("utf-8"))
        elif config_string.startswith("file://"):
            # Spec is a path to a local file
            config_path = config_string[len("file://") :]
            with open(config_path, "r") as f:
                config_dict = json.load(f)
        else:
            # Spec is referring to a configuration from uss_qualifier/configurations
            config_path_parts = [os.path.dirname(__file__)]
            config_path_parts += config_string.split(".")
            config_path = os.path.join(*config_path_parts) + ".json"
            with open(config_path, "r") as f:
                config_dict = json.load(f)
        return ImplicitDict.parse(config_dict, TestConfiguration)
